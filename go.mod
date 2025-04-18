module github.com/TSELab/software_id_core_eda

go 1.23.3

toolchain go1.24.0

require github.com/guacsec/sw-id-core v0.1.1

require (
	github.com/package-url/packageurl-go v0.1.3 // indirect
	golang.org/x/mod v0.24.0 // indirect
)

replace github.com/guacsec/sw-id-core => ../sw-id-core
