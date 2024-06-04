package main

import "flag"

// server [-bind [<ip>]:<port>] [-originHost <originHost>] [-originRealm <originRealm>]
type CommandLineArguments struct {
	Bind             string
	OriginHost       string
	OriginRealm      string
	PathToDictionary string
}

func ProcessCommandLineArguments() (*CommandLineArguments, error) {
	cliArgs := &CommandLineArguments{}

	flag.StringVar(&cliArgs.Bind, "bind", "127.0.0.1:3868", "listen address")
	flag.StringVar(&cliArgs.OriginHost, "originHost", "server.example.com", "asserted OriginHost identity")
	flag.StringVar(&cliArgs.OriginRealm, "originRealm", "example.com", "asserted OriginRealm identity")
	flag.StringVar(&cliArgs.PathToDictionary, "dictionary", "./dictionary.yaml", "path to a Diameter dictionary yaml file")

	flag.Parse()

	return cliArgs, nil
}
