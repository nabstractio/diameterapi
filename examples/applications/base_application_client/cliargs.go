package main

import "flag"

// client [-connect [<ip>]:<port>] [-originHost <originHost>] [-originRealm <originRealm>] [-dictionary /path/to/dictionary]
type CommandLineArguments struct {
	Connect                    string
	OriginHost                 string
	OriginRealm                string
	PathToDictionary           string
	NumberOfSessionsToGenerate uint
}

func ProcessCommandLineArguments() (*CommandLineArguments, error) {
	cliArgs := &CommandLineArguments{}

	flag.StringVar(&cliArgs.Connect, "connect", "127.0.0.1:3868", "peer to connect to")
	flag.StringVar(&cliArgs.OriginHost, "originHost", "client.example.com", "asserted OriginHost identity")
	flag.StringVar(&cliArgs.OriginRealm, "originRealm", "example.com", "asserted OriginRealm identity")
	flag.StringVar(&cliArgs.PathToDictionary, "dictionary", "./dictionary.yaml", "path to a Diameter dictionary yaml file")
	flag.UintVar(&cliArgs.NumberOfSessionsToGenerate, "sessions", uint(1), "number of credit control sessions to generate")

	flag.Parse()

	return cliArgs, nil
}
