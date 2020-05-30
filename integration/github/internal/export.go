package internal

import (
	"strings"
	"time"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/log"
)

const defaultPageSize = 100

type job func(export sdk.Export, pipe sdk.Pipe) error

func (g *GithubIntegration) checkForRateLimit(export sdk.Export, rateLimit rateLimit) error {
	// check for rate limit
	if rateLimit.ShouldPause() {
		if err := export.Paused(rateLimit.ResetAt); err != nil {
			return err
		}
		// pause until we are no longer rate limited
		log.Debug(g.logger, "rate limited", "until", rateLimit.ResetAt)
		time.Sleep(time.Until(rateLimit.ResetAt))
		log.Debug(g.logger, "rate limit wake up")
		// send a resume now that we're no longer rate limited
		if err := export.Resumed(); err != nil {
			return err
		}
	}
	log.Debug(g.logger, "rate limit detail", "remaining", rateLimit.Remaining, "cost", rateLimit.Cost, "total", rateLimit.Limit)
	return nil
}

func (g *GithubIntegration) queuePullRequestJob(repoOwner string, repoName string, repoID string, cursor string) job {
	return func(export sdk.Export, pipe sdk.Pipe) error {
		log.Info(g.logger, "need to run a pull request job starting from "+cursor, "name", repoName, "owner", repoOwner)
		var variables = map[string]interface{}{
			"first": defaultPageSize,
			"after": cursor,
			"owner": repoOwner,
			"name":  repoName,
		}
		for {
			var result repositoryPullrequests
			if err := g.client.Query(pullrequestPageQuery, variables, &result); err != nil {
				return err
			}
			for _, prnode := range result.Repository.Pullrequests.Nodes {
				pullrequest := prnode.ToModel(export.CustomerID(), repoOwner+"/"+repoName, repoID)
				if err := pipe.Write(pullrequest); err != nil {
					return err
				}
				for _, reviewnode := range prnode.Reviews.Nodes {
					prreview := reviewnode.ToModel(export.CustomerID(), repoID, pullrequest.GetID())
					if err := pipe.Write(prreview); err != nil {
						return err
					}
				}
				if prnode.Reviews.PageInfo.HasNextPage {
					// TODO: queue
				}
			}
			if !result.Repository.Pullrequests.PageInfo.HasNextPage {
				break
			}
			if err := g.checkForRateLimit(export, result.RateLimit); err != nil {
				return err
			}
			variables["after"] = result.Repository.Pullrequests.PageInfo.EndCursor
		}
		return nil
	}
}

// Export is called to tell the integration to run an export
func (g *GithubIntegration) Export(export sdk.Export) error {
	log.Info(g.logger, "export started")
	pipe, err := export.Start()
	if err != nil {
		return err
	}
	org := g.config["organization"]
	var variables = map[string]interface{}{
		"login": org,
		"first": 10,
	}
	jobs := make([]job, 0)
	for {
		var result allQueryResult
		if err := g.client.Query(allDataQuery, variables, &result); err != nil {
			export.Completed(err)
			return nil
		}
		for _, node := range result.Organization.Repositories.Nodes {
			repo := node.ToModel(export.CustomerID())
			if err := pipe.Write(repo); err != nil {
				return err
			}
			for _, prnode := range node.Pullrequests.Nodes {
				pullrequest := prnode.ToModel(export.CustomerID(), node.Name, repo.GetID())
				if err := pipe.Write(pullrequest); err != nil {
					return err
				}
				for _, reviewnode := range prnode.Reviews.Nodes {
					prreview := reviewnode.ToModel(export.CustomerID(), repo.GetID(), pullrequest.GetID())
					if err := pipe.Write(prreview); err != nil {
						return err
					}
				}
				if prnode.Reviews.PageInfo.HasNextPage {
					// TODO: queue
				}
			}
			if node.Pullrequests.PageInfo.HasNextPage {
				tok := strings.Split(node.Name, "/")
				// queue the pull requests for the next page
				jobs = append(jobs, g.queuePullRequestJob(tok[0], tok[1], repo.GetID(), node.Pullrequests.PageInfo.EndCursor))
			}
		}
		// check to see if we are at the end of our pagination
		if !result.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
		if err := g.checkForRateLimit(export, result.RateLimit); err != nil {
			pipe.Close()
			export.Completed(err)
			return nil
		}
		variables["after"] = result.Organization.Repositories.PageInfo.EndCursor
	}

	// now cycle through any pending jobs after the first pass
	for _, job := range jobs {
		if err := job(export, pipe); err != nil {
			pipe.Close()
			export.Completed(err)
			return nil
		}
	}

	// finish it up
	if err := pipe.Close(); err != nil {
		return err
	}
	export.Completed(nil)
	return nil
}
