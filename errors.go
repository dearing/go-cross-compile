package main

const (
	NoError                 = iota // 0 - no error
	ErrorUnknown                   // 1 - unknown error
	ErrorConfigFileNotFound        // 2 - config file not found
	ErrorReadConfig                // 3 - error reading config file
	ErrorSrcDirNotFound            // 4 - source directory not found
	ErrorOutDirNotFound            // 5 - output directory not found
	ErrorInitConfig                // 6 - error initializing config
	ErrorOpenArtifact
	ErrorGoBuild
	ErrorMD5SumFile
	ErrorSHA1SumFile
	ErrorSHA256SumFile
	ErrorSHA512SumFile
	ErrorZipFile
)
