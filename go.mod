module github.com/cidverse/go-vcs

go 1.20

require (
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/go-git/go-git/v5 v5.7.0
	github.com/gosimple/slug v1.13.1
	github.com/rs/zerolog v1.29.1
	github.com/sergi/go-diff v1.1.0
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20230528122434-6f98819771a1 // indirect
	github.com/acomagu/bufpipe v1.0.4 // indirect
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.4.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/skeema/knownhosts v1.1.1 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/tools v0.9.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

exclude (
	// 1.2.0 and newer is broken, see https://github.com/sergi/go-diff/issues/123
	github.com/sergi/go-diff v1.2.0
	github.com/sergi/go-diff v1.3.1
)
