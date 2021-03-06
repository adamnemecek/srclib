package src

import (
	"log"
	"net/http"
	"os"

	"github.com/sourcegraph/httpcache"
	"github.com/sourcegraph/httpcache/diskcache"
	"github.com/sqs/go-flags"
	client "sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/task2"
)

var CLI = flags.NewNamedParser("src", flags.Default)

// GlobalOpt contains global options.
var GlobalOpt struct {
	Verbose bool `short:"v" description:"show verbose output"`
}

func init() {
	CLI.LongDescription = "src builds projects, analyzes source code, and queries Sourcegraph."
	CLI.AddGroup("Global options", "", &GlobalOpt)
}

// TODO(sqs): add base URL flag for apiclient
var (
	httpClient = http.Client{Transport: httpcache.NewTransport(diskcache.New("/tmp/srclib-cache"))}
	apiclient  = client.NewClient(&httpClient)
)

var (
	absDir string
)

func init() {
	var err error
	absDir, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
}

func Main() {
	log.SetFlags(0)
	log.SetPrefix("")
	defer task2.FlushAll()

	if _, err := CLI.Parse(); err != nil {
		os.Exit(1)
	}
}

type Subcommand struct {
	Name        string
	Description string
	Run         func(args []string)
}

var Subcommands = []Subcommand{
//	{"tool", "run a tool", toolCmd},
// {"make", "make a repository", make_},
// {"makefile", "print the Makefile and exit", makefile},
//	{"scan", "scan a repository for source units", scanCmd},
// {"config", "validate and print a repository's configuration", config_},
// {"list-deps", "list a repository's raw (unresolved) dependencies", listDeps},
// {"resolve-deps", "resolve a repository's raw dependencies", resolveDeps},
// {"graph", "analyze a repository's source code for definitions and references", graph_},
// {"blame", "blame a source unit's source files to determine commit authors", blame},
// {"authorship", "determine authorship of a source unit's defs and refs", authorship_},
// {"data", "list repository data", dataCmd},
// {"upload", "upload a previously generated build", uploadCmd},
// {"fetch", "downloads build data files from the server", fetchCmd},
// {"person-refresh-profile", "refresh a person's profile", personRefreshProfile},
// {"person-compute-stats", "update a person's stats", personComputeStats},
// {"repo-create", "create a repository (API)", repoCreate},
// {"repo-refresh-profile", "sync repository data", repoRefreshProfile},
// {"repo-refresh-vcs-data", "fetch repository VCS data", repoRefreshVCSData},
// {"repo-compute-stats", "update and print repository stats", repoComputeStats},
// {"build", "create a new build for a repository (API)", build_},
// {"build-queue", "display the build queue (API)", buildQueue},
// {"test", "run tests", testCmd},
//	{"info", "show info about enabled capabilities", infoCmd},
}
