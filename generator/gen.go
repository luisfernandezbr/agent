package generator

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

var _template_readme_md_tmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x5c\xcd\x41\x0a\xc2\x30\x10\x85\xe1\xfd\x9c\xe2\x41\xd6\x7a\x00\x77\x62\x37\x82\x98\x22\xf1\x00\x53\x18\xec\x60\x4c\x24\x9d\x54\xa4\xf4\xee\x42\x11\x0a\xdd\xfe\xbc\xc7\xe7\x1c\xa6\x09\xfb\xa0\x16\xe5\xc4\x83\x5c\xf9\x25\x98\x67\xb4\x9a\xde\x59\x93\xe1\x9c\x4c\x1e\x85\x4d\x73\x22\x72\xce\xc1\x8f\x52\x46\x95\x0f\x51\xf0\x8d\x3f\x20\x70\x7c\x82\xbb\x5c\x0d\xdf\x5c\x0b\x74\x7b\x38\x56\xeb\x73\x21\xda\x2d\x52\x5b\xbb\xa8\x43\x2f\xe5\x2f\x6d\xf3\xfd\x76\x59\x6b\xc3\xb6\x6c\x7e\x01\x00\x00\xff\xff\x55\xc9\x8c\xa1\xa6\x00\x00\x00")

func template_readme_md_tmpl() ([]byte, error) {
	return bindata_read(
		_template_readme_md_tmpl,
		"template/README.md.tmpl",
	)
}

var _template_app_gitignore = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x64\x8c\x41\x6a\x04\x21\x10\x45\xf7\x9e\x42\x98\x5d\x60\x74\x9f\x75\x6e\xd0\x07\x18\x6c\xfd\x63\x1b\xb4\x4a\xca\xb2\x21\xb7\x0f\x66\x12\x32\x30\x9b\xa2\xde\xfb\x9f\x7f\xb1\x1b\x60\x0f\xd5\x3e\xde\xbd\x3f\x50\xbb\xcb\x45\x8f\xb9\xbb\xc8\xcd\x07\xd1\x12\x2b\x86\x2f\x99\x58\x0a\xe5\xeb\xbd\x2c\xb4\x77\x16\xdb\x58\x60\xc3\xce\x53\xed\x5f\x6c\x7f\x62\x67\xcc\xc5\x26\x74\x50\x02\xc5\x82\x61\x3c\x71\xc2\xad\x71\x9a\x75\x91\xeb\xd4\xcd\x3a\xee\x73\xac\xae\x62\x68\xa1\x6c\x7c\xe4\x13\x12\x32\x96\xec\xc2\x69\x46\x2d\x4c\xc6\xef\xb3\xd4\xb4\x64\x2b\x23\x1a\xf7\xb1\xdd\x36\x65\x81\x71\xa0\xd3\x55\x8e\xa1\x3e\xde\x84\x13\x95\x7b\x03\xe9\xb3\x5e\xfb\xcf\xfc\x3f\xfd\x6b\x0d\xf5\x76\x4d\xd8\x67\x76\x95\xf3\x9b\xf9\x0a\x42\x2f\x0c\x11\x96\x07\x7f\x07\x00\x00\xff\xff\x45\xd9\x06\x74\x36\x01\x00\x00")

func template_app_gitignore() ([]byte, error) {
	return bindata_read(
		_template_app_gitignore,
		"template/app/.gitignore",
	)
}

var _template_app_readme_md_tmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x5c\x8d\x41\x0a\xc2\x30\x10\x45\xf7\x73\x8a\x0f\x59\xeb\x01\xdc\x89\xdd\x14\xc4\x16\x69\x0e\x30\x85\xc1\x0e\xc6\x44\xd2\x49\x45\x4a\xef\x2e\x54\xa1\xe0\xf6\xf1\x1e\xcf\x39\xcc\x33\xf6\x9d\x5a\x90\x13\x8f\x72\xe1\x87\x60\x59\xd0\x6a\x7c\x26\x8d\x06\x5f\xa3\x8e\x26\xb7\xcc\xa6\x29\x12\x39\xe7\xd0\x4c\x92\x27\x95\x17\x51\xd7\x54\xcd\x01\x1d\x87\x3b\xb8\x4f\xc5\xf0\x4e\x25\x43\xb7\x00\xbe\xfe\x36\xc7\x62\x43\xca\x44\xbb\xf5\xd7\x96\x3e\xe8\x38\x48\xfe\xfd\xfe\xb1\xbf\x9e\x37\x5a\xb1\xad\xce\x27\x00\x00\xff\xff\xf8\xfd\xe5\xe1\xac\x00\x00\x00")

func template_app_readme_md_tmpl() ([]byte, error) {
	return bindata_read(
		_template_app_readme_md_tmpl,
		"template/app/README.md.tmpl",
	)
}

var _template_app_config_overrides_js = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x54\xcc\xc1\x0d\x42\x21\x0c\x06\xe0\xb3\x9d\xa2\xb7\x07\xc9\xd3\x05\x08\x1b\xb0\x04\xa1\xff\x81\x44\xac\xb6\x60\x8c\xc6\xdd\xbd\x99\xbc\x05\xbe\xa6\x37\x9f\xfc\x61\x7d\xc2\xac\x0b\x76\xae\x22\x05\xee\x45\xab\xc0\xf8\xcb\x99\x0d\x8f\xd5\x0d\x61\x6b\xcb\xa7\x8e\xfe\xc6\xb9\x59\xdd\x62\x22\x1a\x2a\xeb\x8a\x0b\x5e\x77\xb5\xe9\x9c\xff\x4e\xa0\xd3\x01\x0a\x71\xa7\x98\x7e\x01\x00\x00\xff\xff\xbd\x27\x54\x77\x6e\x00\x00\x00")

func template_app_config_overrides_js() ([]byte, error) {
	return bindata_read(
		_template_app_config_overrides_js,
		"template/app/config-overrides.js",
	)
}

var _template_app_package_json_tmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x74\x93\xcd\x6e\xdb\x30\x10\x84\xef\x7e\x0a\x42\x40\x6f\xd1\xda\x94\xdc\x04\xc9\xa1\x08\x90\x53\x81\xa2\x2f\x50\x34\x05\x2d\xae\x1d\x06\x14\x29\xec\xd2\x76\x5c\x41\x7d\xf6\x42\x94\xac\x9f\xd8\xbe\xee\x37\x33\x9c\x95\xbd\xf5\x42\x88\xc4\xa9\x12\x93\x27\x91\x3c\xd7\x35\x7c\xd7\xe8\x82\xd9\x1a\xa4\xa6\x59\xd6\x35\xfc\xf0\x47\xa4\x17\xc5\xf8\x53\x95\xd8\x34\xc9\x5d\x6b\x38\x20\xb1\xf1\xae\xf5\x48\x58\xc1\xaa\x9b\x56\x64\x0e\x2a\xb4\x49\x81\xf6\x18\x47\x1a\x2b\x74\x1a\x5d\x61\x90\x93\x27\xd1\xbe\x26\x44\xf2\x1c\x90\x83\x71\xbb\xd4\x9a\x0d\x29\x3a\x2d\xdf\x91\x43\xaa\x7d\xd9\x06\xbe\xae\x21\x83\x75\x4c\xbc\x26\x25\x54\x45\x88\xba\x47\xc8\x21\xbb\xa9\xdb\x33\x52\x8a\x07\x74\x9d\xf8\x01\xe4\x54\x7c\xaa\x90\xe3\xab\x11\x66\xeb\x61\x89\x91\x3a\xaf\xe3\x47\x79\x95\xd9\x15\x3a\xd6\x90\xf7\xf0\x78\x1d\x0f\x1b\xcd\x25\x05\x73\x6a\xbd\xd2\x48\x11\xe6\xd3\xf4\x62\xcf\xc1\x97\xe6\x2f\xa6\x05\xa9\xce\x3b\xc5\x16\x99\x7b\x93\x5c\x41\x3e\x1d\x4f\x23\xbf\x4e\x3d\xb3\xa6\x32\x07\x39\x03\xa9\xaa\xaa\x94\xf0\x68\x08\x75\xf7\x2d\x40\xc2\xfd\x5c\x32\x59\xe3\xd2\xcf\x05\x99\x2a\xc4\x56\x39\xac\x47\xca\xe1\x64\x71\x5a\x4a\x82\x1c\x1b\xc7\xaf\x14\x9d\x2d\xfb\x97\xc3\x03\x64\xc9\x42\x88\x26\xfe\x6b\xc6\xcc\xfa\x1c\xa6\x28\x2a\x2f\x3a\x8b\x0e\xf5\xb1\x9b\xbd\xb1\xfa\xba\xae\x43\xe7\xe7\xfb\x5f\xfe\x52\x16\x49\xaf\xc2\x77\x2c\x6e\xc8\x3a\x34\x14\x46\xb6\xc6\x85\x17\xef\xb6\x66\x37\xb6\xc6\x8f\x80\x4e\xf3\x2c\x61\xf4\x6c\xc8\x1f\x19\x89\xad\x89\x5d\x7a\x4f\x45\x5e\xef\x8b\xd0\x1d\xd7\xaf\x38\x13\x22\xf9\xb6\x82\xec\x4b\x5f\xab\xbd\x57\x1f\x84\x46\xa5\xe7\x13\x5f\xfd\x29\x8d\x33\x42\x59\x9b\xc4\xf9\xef\x7e\x0f\x8d\x07\xb4\xbe\x2a\xbb\x5b\x18\x42\xad\xe2\x20\xa4\x28\xde\xc8\x97\x28\xce\x27\x7d\xf7\x09\x6f\x0d\xe1\xd6\x7f\xdc\xe4\xac\xb6\x8a\xcc\x80\xbb\x87\x87\x25\xdf\x7c\x89\x95\xda\xc5\x43\x82\x65\xb2\x68\x16\xff\x03\x00\x00\xff\xff\x22\x69\x6d\xa8\x71\x04\x00\x00")

func template_app_package_json_tmpl() ([]byte, error) {
	return bindata_read(
		_template_app_package_json_tmpl,
		"template/app/package.json.tmpl",
	)
}

var _template_app_public_index_html = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\x90\xbf\x4e\xc4\x30\x0c\x87\xf7\x7b\x0a\x93\x99\x12\xb1\x31\x38\x5d\x80\x85\x05\x24\x58\x6e\xf4\x25\x86\x58\x4a\x9d\x2a\x75\x73\xe2\xed\x51\xaf\x07\x62\xf2\x9f\x9f\xfc\x7d\x92\xf1\xe6\xe9\xf5\xf1\xe3\xf8\xf6\x0c\xd9\xa6\x32\x1e\x70\x2b\x50\x48\xbf\x82\x63\x75\xe3\x01\x00\x33\x53\xda\x1a\x00\x9c\xd8\x08\x62\xa6\xb6\xb0\x05\xb7\xda\xe7\xf0\xe0\xc0\xff\x0f\x95\x26\x0e\xae\x0b\x9f\xe7\xda\xcc\x41\xac\x6a\xac\x16\xdc\x59\x92\xe5\x90\xb8\x4b\xe4\xe1\x32\xdc\x82\xa8\x98\x50\x19\x96\x48\x85\xc3\xfd\x15\x85\xfe\xd7\x88\xa7\x9a\xbe\xaf\x74\xad\x4b\x6c\x32\xdb\x78\xac\x2b\x28\x73\x02\xab\xc0\x4a\xa7\xc2\xf0\x42\x9d\xde\x2f\xe9\xb6\x6c\xab\x82\x65\x59\x80\xe6\xf9\x0e\xfd\xdf\xe1\xce\x49\xd2\x41\x52\x70\xad\x56\x73\x23\xfa\x24\x7d\x97\xee\x2e\xf4\xfb\x23\x7e\x02\x00\x00\xff\xff\x19\x69\x1b\xad\x19\x01\x00\x00")

func template_app_public_index_html() ([]byte, error) {
	return bindata_read(
		_template_app_public_index_html,
		"template/app/public/index.html",
	)
}

var _template_app_public_robots_txt = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x52\x56\xc8\x28\x29\x29\x28\xb6\xd2\xd7\x2f\x2f\x2f\xd7\x2b\xca\x4f\xca\x2f\x29\x2e\xa9\x28\xd1\xcb\x2f\x4a\xd7\x47\xf0\x32\x4a\x72\x73\xb8\x42\x8b\x53\x8b\x74\x13\xd3\x53\xf3\x4a\xac\x14\xb4\xb8\x5c\x32\x8b\x13\x73\x72\xf2\xcb\xad\xb8\x00\x01\x00\x00\xff\xff\x55\x9d\xd1\x4a\x43\x00\x00\x00")

func template_app_public_robots_txt() ([]byte, error) {
	return bindata_read(
		_template_app_public_robots_txt,
		"template/app/public/robots.txt",
	)
}

var _template_app_src_app_tsx_tmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x53\x41\x8f\xdb\x2c\x14\x3c\xc3\xaf\x78\x5a\x7d\x5a\x67\xa5\xfd\xec\x9e\xed\x38\xea\xaa\xa7\x95\xa2\x76\x95\x26\xa7\xaa\x07\xd6\x3c\xc7\x68\x31\x20\xc0\xc9\x56\x16\xff\xbd\x02\xc7\x8d\xb3\xed\x0d\x0f\x33\xc3\xbc\x31\x88\xde\x68\xeb\x61\x87\xac\xf1\xd0\x5a\xdd\x43\x66\xe3\x3a\xab\xe8\x65\x6b\x84\xef\xa2\x1f\x24\xf3\xda\x3e\x2b\xe7\x99\x94\x68\x1f\xe1\x59\x79\x3c\x5a\xe6\x85\x56\x10\x2e\xc2\xcf\x46\x28\xe3\x0b\x76\x44\xe5\xf3\x33\xbe\x3a\xfe\x76\xb5\x59\x08\x0e\xcf\x17\x41\x5e\x88\x2b\x9a\x55\x94\xb6\x83\x6a\x92\xe5\x93\x31\xab\x07\x18\x29\x29\x0a\x68\x3a\x6c\xde\xc0\x6b\x70\x88\x20\x5a\x38\x23\x30\x8b\x60\x07\xa5\x84\x3a\x82\xd4\x0d\x93\xc0\x14\x07\x85\xc8\x23\xcf\x0e\x0a\x84\x02\x37\xc5\x8e\x76\xbd\xe6\x48\x89\x68\x61\x75\x16\x8a\xeb\x33\xd4\x75\x0d\xd3\x32\x37\xcc\xa2\xf2\x70\x7f\x3f\x03\xd1\x30\xaa\xf2\xce\x62\x9b\x0b\xc5\xf1\xfd\x5b\xbb\xca\xd2\x39\x9d\x76\x3e\x7b\x80\x0d\x7c\x4a\xe9\x48\xa3\x95\xf3\xb0\x98\xa2\xbc\x69\xa6\x4e\x24\xa2\x58\x8f\x25\x64\xe3\x08\xf9\xcb\xf0\x2a\x85\xeb\xd0\x7e\x65\x3d\x42\x08\xd9\x63\x24\x70\x74\x8d\x15\x66\x32\xc8\xf6\x9d\x70\x20\x1c\xf8\x0e\x21\x6a\xf6\xc2\x4b\xfc\xc2\x1c\x5e\x34\xcb\xf3\xa0\xd5\x16\x5e\x84\x32\x5a\x28\x3f\x99\x79\x76\x74\x25\xfc\x88\x4b\x32\x8e\xff\x83\x65\xea\x88\xf0\xdf\x89\x49\x28\x6b\xc8\x17\xf9\xf6\xbf\x0c\x3a\x08\x21\x51\x49\xcc\x97\x58\x73\x2a\x80\xa4\x47\xc5\x27\xca\xcf\x84\x8a\xcb\x25\xe0\x25\xb4\x4c\x3a\x4c\xa0\xc5\x36\x9a\xa5\x21\xf3\xad\x3e\xa3\x9d\xf3\xce\x66\xa2\x49\xb3\x4d\x1f\x66\x6e\xa1\x9c\x0a\xba\x36\x74\x5b\xd0\x2c\x26\xec\xc4\x3c\xb3\x1f\x18\x4f\x09\xfc\xc3\x19\xac\xfc\x40\x38\xec\xb6\x21\x64\x71\x33\x24\xca\x20\x0e\xbb\x6d\xf9\xcf\xff\x4c\x09\x09\x15\x8d\x83\xf8\xc1\x2a\x58\xff\x7d\xe7\x97\xad\xd7\xe3\xe2\x23\x40\xb1\xa9\x28\x09\x74\x16\xaf\x28\x21\x6b\x2e\x4e\xd0\x48\xe6\x5c\x9c\xa3\xbe\x7b\x32\xe6\x6e\x13\x33\xac\x6f\x1f\x42\x11\xc1\x75\xc1\xc5\x69\x43\xc9\x43\x45\x03\xa5\xf8\x9e\x1e\x0c\xc7\x96\x0d\xd2\xc7\xa7\x50\xd1\xdf\x01\x00\x00\xff\xff\x74\x25\x9b\x65\xa7\x03\x00\x00")

func template_app_src_app_tsx_tmpl() ([]byte, error) {
	return bindata_read(
		_template_app_src_app_tsx_tmpl,
		"template/app/src/App.tsx.tmpl",
	)
}

var _template_app_src_index_tsx_tmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x8f\x41\x4e\xeb\x30\x10\x86\xd7\xf6\x29\xac\x6e\xdc\x4a\xef\xd9\x07\x48\xa8\x08\x85\x05\x12\x85\x0a\xb8\x40\x1a\x4f\x8b\x45\x9d\x19\x4d\xa6\x94\x2a\xca\xdd\x51\xa2\x54\xa9\x04\xbb\xb1\xbf\xdf\xdf\xef\x89\x89\x90\xc5\xbc\x42\x59\x89\xd9\x31\x26\x63\xb9\x9f\x6d\xa6\xaf\xd1\xfd\xcb\xfa\x9a\xfe\x0f\x98\xa6\x44\x6b\x0a\xa2\x15\xd6\x02\xdf\x62\xba\x31\x77\x4b\xb1\x26\xf1\xe5\x1e\x6a\x71\x27\xd8\x36\xe1\x73\x7a\x51\x10\x8d\x31\xe7\x0b\x22\x9b\x69\x7d\xa9\x71\x0c\x75\x00\x9e\x6b\x95\x0f\x57\xee\x4d\x38\x56\xb2\xc6\x00\x4b\xad\x54\x3e\x55\xb9\x0d\xe3\x57\x0c\xc0\x86\x8e\xdb\x43\x6c\x3e\x80\x6f\x66\x6d\xeb\x36\x97\xd3\x73\x99\xa0\xeb\x66\x86\x61\xf7\x7e\x26\x18\xe0\x13\x9e\x80\x57\x65\x03\x23\xec\x9d\x83\xd4\xf8\x41\xef\xff\xf0\x2f\xb5\xca\xfd\xaf\xcf\xfc\xd3\x2a\x60\x75\x4c\xfd\x7e\x7b\x90\x87\x03\xf4\xe3\xdd\xf9\x31\xcc\x2d\x23\x8a\x5d\xe8\x45\xa6\x7f\x02\x00\x00\xff\xff\x23\x1f\x05\x22\x60\x01\x00\x00")

func template_app_src_index_tsx_tmpl() ([]byte, error) {
	return bindata_read(
		_template_app_src_index_tsx_tmpl,
		"template/app/src/index.tsx.tmpl",
	)
}

var _template_app_src_integration_tsx = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x54\x8e\xb1\x4e\xc4\x30\x10\x44\x6b\xef\x57\x6c\x97\xbb\x26\xf7\x01\xbe\xa4\xa1\x4a\x43\x81\x44\x41\x69\xc5\x9b\x60\xc9\xf6\xa2\xdd\x75\x04\x8a\xf2\xef\x08\x48\xc1\x75\x33\x7a\xd2\x9b\x49\xe5\x83\xc5\xf0\x85\xc2\x6c\xb8\x08\x17\xec\xe4\x27\x77\x1e\x4e\xa4\xf6\x95\x49\x4f\xd6\xdf\xfe\x6a\x5f\x38\xb6\x4c\x7d\x26\xd5\xce\x03\xcc\x5c\xd5\x70\xaa\x46\xab\x04\x4b\x5c\x71\xc0\xcb\x15\x87\x11\x77\x70\x42\xd6\xa4\xe2\x05\x9c\xbb\xc7\xb4\xe1\x9c\x83\xea\x73\x28\x34\xec\xa7\xed\x89\xab\x85\x54\x49\x8e\xf1\x8d\x9b\x60\xfa\x27\x7a\x9d\x70\x65\x52\x7c\x27\xa1\xfb\x2d\xa6\x6d\x04\x77\xf5\x70\x78\x00\xfa\xfc\x7d\x18\x69\x09\x2d\x3f\xcc\xfb\xef\x00\x00\x00\xff\xff\xa7\x4c\x64\x80\xd8\x00\x00\x00")

func template_app_src_integration_tsx() ([]byte, error) {
	return bindata_read(
		_template_app_src_integration_tsx,
		"template/app/src/integration.tsx",
	)
}

var _template_app_src_react_app_env_d_ts = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x3c\xcb\x31\xaa\x83\x50\x10\x05\xd0\xda\x59\xc5\xc5\x46\xf8\xf0\xf3\x7a\x4d\xb2\x91\x90\xe2\x31\x5e\x83\xe4\x45\x65\x66\x84\x88\xb8\xf7\x54\x49\x77\x9a\x93\x12\x12\xce\xc6\x81\xc6\x49\x89\xd8\x16\xfa\xa5\x36\x66\x8d\x7f\x57\x1b\x97\xf0\x1a\xe9\x2a\xd2\x53\x4b\x36\xe2\x35\xf7\x6b\x21\x9a\xbf\x53\xa1\x7b\x83\x5d\x2a\x9d\x27\x0f\x68\xc9\xee\xf4\x16\x3b\x6e\x4f\x6e\x2d\x3c\x6c\x9c\x1e\xf7\x2f\x70\x74\x52\xf1\xbd\xcc\x16\xe8\x39\xe4\xb5\xfc\x4e\x27\x87\x7c\x02\x00\x00\xff\xff\xc1\x76\xa2\x73\x8a\x00\x00\x00")

func template_app_src_react_app_env_d_ts() ([]byte, error) {
	return bindata_read(
		_template_app_src_react_app_env_d_ts,
		"template/app/src/react-app-env.d.ts",
	)
}

var _template_app_src_styles_module_less = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\xc8\xcc\x2d\xc8\x2f\x2a\x51\x50\xaf\x73\x28\xc8\xcc\x2b\x28\xd1\x2f\xcd\x4c\xd6\xcb\x4b\xad\x28\xd1\x0f\x2e\xa9\xcc\x49\x2d\xd6\x2f\x06\x53\x7a\x39\xa9\xc5\xc5\xea\xd6\x5c\x5c\x7a\xce\xf9\x79\x25\x89\x99\x79\xa9\x45\x0a\xd5\x5c\x9c\xb9\x89\x45\xe9\x99\x79\x56\x0a\x06\xd6\x5c\x9c\x05\x89\x29\x29\x99\x79\xe9\x60\x4e\x2d\x20\x00\x00\xff\xff\xab\xc6\x7c\x0c\x57\x00\x00\x00")

func template_app_src_styles_module_less() ([]byte, error) {
	return bindata_read(
		_template_app_src_styles_module_less,
		"template/app/src/styles.module.less",
	)
}

var _template_app_tsconfig_json = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x5c\x91\xc1\x4e\xf3\x30\x10\x84\xef\x7d\x0a\x6b\xcf\xbf\xfe\x1b\x97\x5e\x0b\x48\xad\x28\x48\x70\x44\x3d\x38\xce\xb6\x5d\x6a\xef\x46\xde\x0d\x14\xa1\xbe\x3b\xb2\x93\x02\xc9\x71\xbe\x89\x66\x36\xe3\xaf\x85\x73\x10\x24\x75\x14\x31\x3f\x75\x46\xc2\x0a\x4b\x57\xb0\x73\x60\x3e\x1f\xd0\x60\xe9\x00\xf5\x06\xfe\x0d\x30\x52\x03\x4b\xf7\x5a\x85\x73\xd0\x4a\x1a\x9d\x41\xfc\x27\xc3\xec\x9b\x88\xbf\x14\x95\xf1\x6c\x50\xe5\x6e\x4c\xf1\x31\xca\xc7\xa6\x74\x59\xee\x71\x84\x7a\xa2\xee\x81\x9a\xd5\x11\xc3\x69\xea\xa0\x6e\xa5\xed\x23\xae\xd9\x30\x4b\x37\x35\x6b\xd6\xcb\x27\xdb\x11\x8d\xc2\x2d\xee\x7d\x1f\x6d\x9d\x3a\xc9\x36\x2f\xb0\x4c\xc1\xa6\x6c\x2f\x39\xe0\x4a\x58\x49\x0d\xd9\x56\x5e\x89\x0f\x6b\xbe\xa7\x88\x8f\x3e\xe1\x2c\x21\xd5\x33\x86\x49\xea\x5f\x4d\xf8\x33\xaa\xc4\xbe\xac\x58\xbe\x60\x69\xaf\x2b\x40\x2e\xce\x3b\x6e\x54\x78\x7b\x8d\xf8\x13\x4b\x2a\xd1\x1b\xb6\x83\x37\xeb\x64\xb9\x4b\x34\xbb\xfa\x4d\xcf\xa5\x22\xa3\x0f\x75\xd9\x4b\xe1\x40\x1c\x62\xdf\xe2\xcf\xfb\x80\xe6\x50\xdc\xdd\xe2\xb2\xf8\x0e\x00\x00\xff\xff\xc9\x20\xaf\x73\xeb\x01\x00\x00")

func template_app_tsconfig_json() ([]byte, error) {
	return bindata_read(
		_template_app_tsconfig_json,
		"template/app/tsconfig.json",
	)
}

var _template_go_mod_tmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xca\xcd\x4f\x29\xcd\x49\x55\xa8\xae\xd6\x0b\xc8\x4e\xaf\xad\xe5\xe2\x4a\xcf\x57\x30\xd4\x33\x34\xe1\xe2\x2a\x4a\x2d\x2c\xcd\x2c\x4a\x55\xd0\xe0\xe2\x4c\xcf\x2c\xc9\x28\x4d\xd2\x4b\xce\xcf\xd5\x2f\xc8\xcc\x2b\x28\xd1\x4f\x4c\x4f\xcd\x2b\xd1\x2f\x33\x01\xe9\x73\xcf\x2c\x09\x49\x04\x69\xd5\xe4\x02\x04\x00\x00\xff\xff\xab\x0b\x56\x3e\x4d\x00\x00\x00")

func template_go_mod_tmpl() ([]byte, error) {
	return bindata_read(
		_template_go_mod_tmpl,
		"template/go.mod.tmpl",
	)
}

var _template_integration_go_tmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x54\x8e\xb1\x6e\xc2\x40\x0c\x86\xe7\xf8\x29\xac\x0c\x15\x19\x38\x2f\x7d\x83\xd2\x81\xa1\xd0\x21\x2f\xe0\x06\x73\xb1\x48\x9c\xe8\xe2\xa0\x4a\xa7\x7b\xf7\x0a\x5a\xa9\xb0\x5a\x9f\xff\xef\x23\xc2\xdd\x11\x0f\xc7\x16\xdf\x77\xfb\x16\xb7\x5b\xf4\x5e\x17\xd4\x05\x19\xa3\x98\x24\x76\x39\xe1\x59\x07\x81\x99\xbb\x0b\x47\xc1\x91\xd5\x00\x74\x9c\xa7\xe4\xb8\x81\xaa\xce\x39\x7c\x5e\x62\x29\xa4\xe6\x92\x8c\x87\x1a\xaa\x3a\xaa\xf7\xeb\x57\xe8\xa6\x91\x66\xb5\xd9\x89\xa3\x98\xd3\xf5\x95\xd2\x6a\x26\xa9\x86\x06\x80\x08\xf7\xe6\x12\x13\xbb\x4e\x76\xb3\xae\x8b\x9c\xd0\x27\x94\xef\xfb\xbc\xf7\x82\xfa\x4f\xc0\x95\xd3\xf3\xc7\x9f\x31\xe4\x1c\x5a\xf5\x41\xde\x78\x91\x03\x8f\x52\xca\x03\x06\x70\x5e\xad\xbb\x87\x6f\x1a\xcc\x50\xfd\x26\x84\x8f\xdb\xe1\xe5\x01\x6c\xa0\xc0\x4f\x00\x00\x00\xff\xff\xf1\x1d\x67\xea\x12\x01\x00\x00")

func template_integration_go_tmpl() ([]byte, error) {
	return bindata_read(
		_template_integration_go_tmpl,
		"template/integration.go.tmpl",
	)
}

var _template_integration_yaml_tmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x54\x8d\xb1\x6a\x03\x31\x10\x44\x7b\x7d\xc5\x20\xd2\xda\x1f\x20\x48\xe5\x36\x84\x14\xee\xcd\xc6\xda\x3b\x2f\xe8\x56\x87\xb4\x76\x08\x62\xff\x3d\xdc\xe1\xc2\xe9\x06\xde\xcc\x3c\xa5\x85\x13\xe2\x18\xc7\xb3\x58\xe1\x13\x75\xfe\xa4\x85\xdd\x63\x68\x3c\x5d\xec\x77\x7d\xe2\x8f\xfa\xc3\xed\x15\x67\xee\xd7\x26\xab\x49\xd5\x84\x78\xbe\x49\x87\x74\xd8\x8d\x31\x06\xfe\xbf\xc1\x1d\xa2\xc6\x73\xa3\xad\x8e\xa9\x36\x7c\x89\xae\x55\xd4\x62\xa0\x07\x19\xb5\xcb\xbd\x95\x84\x18\xc3\x95\x56\xfa\x96\x22\x26\xdc\x53\x18\xe3\x80\x46\x3a\x33\xde\x1e\x54\x90\xde\x71\x3c\xbd\x14\xe0\x1e\x80\xc3\xa6\xdc\xb9\xfb\xbe\x60\xcd\xee\x41\xb4\x1b\x95\xb2\x3b\x53\x00\x96\x9a\xb7\x4b\x60\x5b\x5c\x4b\xbd\xe7\x67\xee\x5c\xa6\x85\x94\x66\xce\xe1\x2f\x00\x00\xff\xff\xfe\x51\xc2\x86\x11\x01\x00\x00")

func template_integration_yaml_tmpl() ([]byte, error) {
	return bindata_read(
		_template_integration_yaml_tmpl,
		"template/integration.yaml.tmpl",
	)
}

var _template_internal_root_go_tmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x57\x4d\x6f\xdb\x38\x10\x3d\x5b\xbf\x62\xe0\x93\x5c\x78\xed\xcb\x9e\x0a\xf8\x50\x24\x05\x36\x40\x5b\x2c\x9a\x60\xf7\xb0\x58\x14\x34\x35\x92\xb8\xa1\x48\x96\x1c\xda\x0d\x82\xfc\xf7\x05\xbf\x6c\x39\x51\x1a\xfb\x50\x9f\xa4\x11\xe7\xcd\xbc\x37\xc3\x21\x6d\x18\xbf\x67\x1d\x82\x50\x84\x56\x31\x59\x55\x62\x30\xda\x12\xd4\xd5\x6c\xde\x09\xea\xfd\x76\xc5\xf5\xb0\x36\x42\x19\x5a\xb3\x0e\x15\xad\x77\xbf\xaf\x5d\x73\x3f\xaf\x16\x55\xb5\x5e\xc3\xe3\x23\xac\xee\x04\x49\xbc\x62\x0e\xbf\xb0\x01\xe1\xe9\xe9\x46\x11\x76\x96\x91\xd0\x0a\x84\x03\xa6\x22\x7e\xb1\xb4\xda\x4e\x7a\x55\xf4\x60\xf0\x4d\x3c\x47\xd6\x73\x82\xc7\x6a\xc6\xb5\x6a\x45\x07\xe0\x9a\xfb\xd5\x55\x7c\xae\x66\x03\x53\xac\x43\x1b\x6d\x9f\xd3\x73\x35\xb3\xd8\xde\x05\x68\x47\x56\xa8\xae\x7a\xaa\xaa\x1d\xb3\xf0\x2d\x2e\x1a\x43\x6f\xa0\x7e\xf7\x46\xf8\x45\xad\x84\x4c\xc4\x6f\x89\x59\x0a\xf4\x38\x93\x12\x1b\xd8\xf7\xa8\x80\x7a\x3c\xe1\x2a\x1c\xb8\xb0\x4e\xa8\x0e\xbc\xa9\x5a\xaf\x38\xd4\x1d\xbc\x19\x26\xa1\xd7\x52\x77\x85\xcd\xa7\xf8\xb8\x84\xcc\xfa\x48\x7a\x09\x13\xa4\x17\x80\xd6\x06\x9d\xab\x59\xb7\xca\x2e\x9b\xec\x1b\x4c\xc5\x65\x53\x9c\x83\xb1\xe8\xb4\x81\x79\xc8\xef\x93\xde\xa3\x1d\xe5\x37\xaf\x66\x39\x91\x1b\xd5\xea\x9c\xdb\x12\xe6\x85\xe0\x7c\x11\xa4\x26\x6f\x15\x28\x21\x83\xcc\x51\x25\x6d\xce\x11\xa9\xf7\x14\x45\x6a\xf4\x3e\x75\x08\x97\xc8\xd4\x85\x92\x69\xf3\x52\xb1\x91\x12\xaf\x64\xaf\x8d\x99\xcc\x3e\xa4\xff\x51\x59\x2d\xe5\x73\x02\x0c\x14\xee\x4f\x29\x28\x47\x4c\x71\x8c\xed\xde\x34\xd8\x5c\x90\x77\x8a\x51\x1f\x20\x52\x5b\xa6\x97\x51\xf6\x99\xd9\xfb\xcd\x21\x58\x66\x58\x2f\x5e\xa1\x86\x29\x79\xa5\x09\xc4\x60\x24\x0e\xa8\x08\x9b\xc9\x32\x5d\x0b\x37\x08\xe7\x5e\x10\x55\x80\x3f\x84\x8b\x95\x79\x8d\xae\xc5\x41\xef\x2e\x22\x9c\x83\xfd\x02\xc6\x4d\xa6\x71\x0e\xe5\xbf\x71\xfb\x87\xd6\xf7\x2f\x6b\xbb\xc7\x6d\x9f\x3f\x58\xe4\x28\x76\xd8\x80\x56\xb0\xc5\x9e\xc9\x16\x74\xfb\xbc\x7f\x2f\x60\x9e\x63\xd6\x25\x44\x20\x91\x6d\x93\xbc\xf3\xba\xb7\x68\x17\xb8\x73\x68\x7f\xf6\x74\xd8\x75\xa7\xbc\x87\xf2\xc5\xe2\x77\x8f\x8e\x7e\x85\x00\x25\x7a\x7d\x08\x16\x87\x56\x7e\x59\x40\xfd\x6e\xfc\xfe\x15\x9d\xd1\xca\xe1\x32\x69\xb3\x38\x15\xa7\x40\x8c\xd4\xf9\x06\x1b\x48\x0b\xc6\xcc\x97\x23\xfa\x1f\x7f\xc4\xe3\xed\x48\x9e\x34\x10\x4a\xf9\x62\x28\x91\x06\xeb\xf3\x0e\x08\x2e\x97\x6c\xe8\xe8\x50\x27\xbf\x48\x30\x59\x26\x4b\x9c\x56\x8d\x38\x54\xb3\x38\x50\xb1\x09\x5f\x49\x0c\xb8\xfa\xa2\xf7\xaf\xef\xf1\x1c\x24\xb9\xcc\x17\x55\x35\x5b\xaf\xe1\x4f\x61\x10\x06\xef\x08\xb6\x38\x22\xba\xc5\x4e\x8c\x28\x01\x53\x4d\xa9\x30\x30\x30\xc1\x29\x8c\x5e\x87\xaa\x89\xb3\x98\x11\xab\x66\xd1\x7c\x4c\x34\x40\xd7\x39\xcc\x2d\x31\x4a\x23\x0f\xb8\x77\xa4\x87\x30\x79\x0d\x72\xd1\x0a\x1e\x52\x22\x04\xbd\xfd\x0f\x39\x45\x58\xea\x85\x3b\x91\x38\x84\x2f\x7e\x91\x35\x8d\x03\x45\xf0\x12\xe9\x2a\x2f\xbb\xb9\x86\xbd\x90\x12\x72\x71\x43\xd1\x0e\x91\x45\x93\xc3\x60\x29\xd9\x8c\x1f\xdd\x8e\xc0\x47\xac\x03\x7a\x3a\x20\xe3\x55\xe5\x61\x82\x4a\x3a\x37\xfd\xe8\xfe\x12\xb9\x1c\x73\xcf\x27\xec\x28\x46\x34\x44\xfc\x5c\xb6\x6b\xdc\xfa\x6e\xba\x6e\xe9\xf4\x89\x99\xc4\xf7\x52\x9e\x1e\x2d\x46\xeb\x6f\xe3\x5f\xb4\x7c\xf7\x82\x30\xd1\xd7\x83\x11\x12\xed\xfb\xd4\xfd\xa1\x5a\xe9\x29\xea\x99\x1e\x8f\x32\xe4\xf7\x74\x0f\xa8\x7e\xde\x53\xad\x50\xc2\xf5\xd8\xcc\xc3\x60\xcd\xe4\xe7\xcb\xd4\x92\xb7\x42\x71\xac\x73\xd7\x2d\xa6\xe6\xcc\x07\x4f\xfa\x2a\xeb\x86\x2f\x87\x0d\x97\xda\x37\x27\xdd\xd0\x33\x57\x26\x4f\xe9\x56\x60\x9e\xf4\x41\xfd\x8b\xce\x98\x93\xf0\x75\xc0\x19\x5d\x9c\x4e\x3e\x96\xa1\x53\x6e\x53\x53\xa3\xe6\xe8\x7f\xd1\xb0\xf9\x8b\x49\xd1\xe4\x2d\x92\xe9\x6f\xb1\xd5\x16\xcf\xb9\x41\x04\x09\x1a\x24\xb4\x83\x50\x18\xd0\x44\x9b\x0b\x5e\x9a\x75\x17\xe0\xe3\x2e\x7a\x3e\xbc\x38\x53\x60\xac\x36\x68\xe5\x43\xe8\x90\xc1\x2b\xc1\x43\x22\x7b\x41\x7d\x58\x1d\xf0\x9c\xf6\x36\x1c\xbe\x0f\x8e\x70\x58\xc1\x5d\x8f\x60\xd1\x79\x49\x07\xc8\x34\xae\xe2\x86\xdb\x6a\xea\x43\x45\x1a\x94\x62\x87\x36\x0f\xce\x1e\xe1\x83\x31\xab\x80\xf6\x35\x2a\x10\xc6\x06\x0b\x0a\x64\xdf\x40\x5c\x2b\x27\x9a\xe8\xc2\xc0\x79\xce\xd1\xb9\xd6\xcb\x94\x7d\x1a\xe0\xe7\xd7\xb5\x28\x5a\xef\x8a\xb4\xa1\x74\xc5\xba\x80\x3a\x33\x18\x98\xf9\x27\x5d\xf7\xff\x8d\x7f\x6d\x5a\xc6\xf1\xf1\x29\x16\x77\xb2\xc0\x05\xee\xec\xf2\xfe\x1f\x00\x00\xff\xff\x14\xa2\x05\xe8\x39\x0d\x00\x00")

func template_internal_root_go_tmpl() ([]byte, error) {
	return bindata_read(
		_template_internal_root_go_tmpl,
		"template/internal/root.go.tmpl",
	)
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() ([]byte, error){
	"template/README.md.tmpl": template_readme_md_tmpl,
	"template/app/.gitignore": template_app_gitignore,
	"template/app/README.md.tmpl": template_app_readme_md_tmpl,
	"template/app/config-overrides.js": template_app_config_overrides_js,
	"template/app/package.json.tmpl": template_app_package_json_tmpl,
	"template/app/public/index.html": template_app_public_index_html,
	"template/app/public/robots.txt": template_app_public_robots_txt,
	"template/app/src/App.tsx.tmpl": template_app_src_app_tsx_tmpl,
	"template/app/src/index.tsx.tmpl": template_app_src_index_tsx_tmpl,
	"template/app/src/integration.tsx": template_app_src_integration_tsx,
	"template/app/src/react-app-env.d.ts": template_app_src_react_app_env_d_ts,
	"template/app/src/styles.module.less": template_app_src_styles_module_less,
	"template/app/tsconfig.json": template_app_tsconfig_json,
	"template/go.mod.tmpl": template_go_mod_tmpl,
	"template/integration.go.tmpl": template_integration_go_tmpl,
	"template/integration.yaml.tmpl": template_integration_yaml_tmpl,
	"template/internal/root.go.tmpl": template_internal_root_go_tmpl,
}
// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() ([]byte, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"template": &_bintree_t{nil, map[string]*_bintree_t{
		"README.md.tmpl": &_bintree_t{template_readme_md_tmpl, map[string]*_bintree_t{
		}},
		"app": &_bintree_t{nil, map[string]*_bintree_t{
			".gitignore": &_bintree_t{template_app_gitignore, map[string]*_bintree_t{
			}},
			"README.md.tmpl": &_bintree_t{template_app_readme_md_tmpl, map[string]*_bintree_t{
			}},
			"config-overrides.js": &_bintree_t{template_app_config_overrides_js, map[string]*_bintree_t{
			}},
			"package.json.tmpl": &_bintree_t{template_app_package_json_tmpl, map[string]*_bintree_t{
			}},
			"public": &_bintree_t{nil, map[string]*_bintree_t{
				"index.html": &_bintree_t{template_app_public_index_html, map[string]*_bintree_t{
				}},
				"robots.txt": &_bintree_t{template_app_public_robots_txt, map[string]*_bintree_t{
				}},
			}},
			"src": &_bintree_t{nil, map[string]*_bintree_t{
				"App.tsx.tmpl": &_bintree_t{template_app_src_app_tsx_tmpl, map[string]*_bintree_t{
				}},
				"index.tsx.tmpl": &_bintree_t{template_app_src_index_tsx_tmpl, map[string]*_bintree_t{
				}},
				"integration.tsx": &_bintree_t{template_app_src_integration_tsx, map[string]*_bintree_t{
				}},
				"react-app-env.d.ts": &_bintree_t{template_app_src_react_app_env_d_ts, map[string]*_bintree_t{
				}},
				"styles.module.less": &_bintree_t{template_app_src_styles_module_less, map[string]*_bintree_t{
				}},
			}},
			"tsconfig.json": &_bintree_t{template_app_tsconfig_json, map[string]*_bintree_t{
			}},
		}},
		"go.mod.tmpl": &_bintree_t{template_go_mod_tmpl, map[string]*_bintree_t{
		}},
		"integration.go.tmpl": &_bintree_t{template_integration_go_tmpl, map[string]*_bintree_t{
		}},
		"integration.yaml.tmpl": &_bintree_t{template_integration_yaml_tmpl, map[string]*_bintree_t{
		}},
		"internal": &_bintree_t{nil, map[string]*_bintree_t{
			"root.go.tmpl": &_bintree_t{template_internal_root_go_tmpl, map[string]*_bintree_t{
			}},
		}},
	}},
}}
