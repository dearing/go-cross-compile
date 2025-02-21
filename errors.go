package main

const (
	NoError                 = iota // 0 - no error
	ErrorUnknown                   // 1 - unknown error
	ErrorConfigFileNotFound        // 2 - config file not found
	ErrorReadConfig                // 3 - error reading config file
	ErrorSrcDirNotFound            // 4 - source directory not found
	ErrorOutDirNotFound            // 5 - output directory not found
	ErrorInitConfig                // 6 - error initializing config
	ErrorOpenArtifact              // 7 - error opening artifact
	ErrorGoBuild                   // 8 - error building binary
	ErrorMD5SumFile                // 9 - error creating MD5 checksum file
	ErrorSHA1SumFile               // 10 - error creating SHA1 checksum file
	ErrorSHA256SumFile             // 11 - error creating SHA256 checksum file
	ErrorSHA512SumFile             // 12 - error creating SHA512 checksum file
	ErrorZipFile                   // 13 - error creating zip file
)
