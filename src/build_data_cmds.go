//+build off

package src

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"code.google.com/p/rog-go/parallel"

	"github.com/sourcegraph/rwvfs"

	client "sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/buildstore"
)

func dataCmd(args []string) {
	fs := flag.NewFlagSet("data", flag.ExitOnError)
	remote := fs.Bool("remote", true, "show remote data")
	local := fs.Bool("local", true, "show local data")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `usage: `+Name+` data [options]

Lists available repository data.

The options are:
`)
		fs.PrintDefaults()
		os.Exit(1)
	}
	fs.Parse(args)
	if fs.NArg() != 0 {
		fs.Usage()
	}

	repo, err := OpenRepo(Dir)
	if err != nil {
		log.Fatal(err)
	}

	if *remote {
		log.Println("===================== REMOTE")
	}

	remoteFiles, resp, err := apiclient.BuildData.List(client.RepositorySpec{URI: string(repo.URI()), CommitID: repo.CommitID}, nil)
	if err != nil {
		if hresp, ok := resp.(*client.HTTPResponse); hresp != nil && ok && hresp.StatusCode == http.StatusNotFound {
			log.Println("No remote build data found.")
		} else {
			log.Fatal(err)
		}
	}

	if *remote {
		if remoteFiles != nil {
			PrintJSON(remoteFiles, "")
		}
		log.Println("============================")
	}

	repoStore, err := buildstore.NewRepositoryStore(repo.RootDir)
	if err != nil {
		log.Fatal(err)
	}

	localFiles, err := repoStore.AllDataFiles()
	if err != nil {
		log.Fatal(err)
	}

	if *local {
		log.Println("===================== LOCAL")
		PrintJSON(localFiles, "")
		log.Println("============================")
	}
}

func fetchCmd(args []string) {
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `usage: `+Name+` fetch [options]

Fetches build data for a repository from Sourcegraph.

The options are:
`)
		fs.PrintDefaults()
		os.Exit(1)
	}
	fs.Parse(args)
	if fs.NArg() != 0 {
		fs.Usage()
	}

	repo, err := OpenRepo(Dir)
	if err != nil {
		log.Fatal(err)
	}

	localFiles, resp, err := apiclient.BuildData.List(client.RepositorySpec{URI: string(repo.URI()), CommitID: repo.CommitID}, nil)
	if err != nil {
		if hresp, ok := resp.(*client.HTTPResponse); hresp != nil && ok && hresp.StatusCode == http.StatusNotFound {
			log.Println("No remote build data found.")
			return
		} else {
			log.Fatal(err)
		}
	}

	repoStore, err := buildstore.NewRepositoryStore(repo.RootDir)
	if err != nil {
		log.Fatal(err)
	}

	par := parallel.NewRun(15)
	for _, file_ := range localFiles {
		file := file_
		par.Do(func() error {
			fetchFile(repoStore, string(repo.URI()), file)
			return nil
		})
	}
	par.Wait()
}

func fetchFile(repoStore *buildstore.RepositoryStore, repoURI string, fi *buildstore.BuildDataFileInfo) {
	path := repoStore.FilePath(fi.CommitID, fi.Path)

	fileSpec := client.BuildDataFileSpec{
		Repo: repoURI,
		Rev:  fi.CommitID,
		Path: fi.Path,
	}

	kb := float64(fi.Size) / 1024
	if GlobalOpt.Verbose {
		log.Printf("Fetching %s (%.1fkb)", path, kb)
	}

	data, _, err := apiclient.BuildData.Get(fileSpec)
	if err != nil {
		log.Fatal(err)
	}

	if GlobalOpt.Verbose {
		log.Printf("Fetched %s (%.1fkb)", path, kb)
	}

	err = rwvfs.MkdirAll(repoStore, filepath.Dir(path))
	if err != nil {
		log.Fatal(err)
	}

	f, err := repoStore.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		log.Fatal(err)
	}

	if GlobalOpt.Verbose {
		log.Printf("Saved %s", path)
	}
}

func uploadCmd(args []string) {
	fs := flag.NewFlagSet("upload", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `usage: `+Name+` upload [options]

Uploads build data for a repository to Sourcegraph.

The options are:
`)
		fs.PrintDefaults()
		os.Exit(1)
	}
	fs.Parse(args)
	if fs.NArg() != 0 {
		fs.Usage()
	}

	repo, err := OpenRepo(Dir)
	if err != nil {
		log.Fatal(err)
	}

	repoStore, err := buildstore.NewRepositoryStore(repo.RootDir)
	if err != nil {
		log.Fatal(err)
	}

	localFiles, err := repoStore.AllDataFiles()
	if err != nil {
		log.Fatal(err)
	}

	par := parallel.NewRun(15)
	for _, file_ := range localFiles {
		file := file_
		par.Do(func() error {
			uploadFile(repoStore, file, string(repo.URI()))
			return nil
		})
	}
	par.Wait()
}

func uploadFile(repoStore *buildstore.RepositoryStore, file *buildstore.BuildDataFileInfo, repoURI string) {
	path := repoStore.FilePath(file.CommitID, file.Path)

	fileSpec := client.BuildDataFileSpec{
		Repo: repoURI,
		Rev:  file.CommitID,
		Path: file.Path,
	}

	fi, err := repoStore.Stat(path)
	if err != nil || !fi.Mode().IsRegular() {
		if GlobalOpt.Verbose {
			log.Printf("upload: skipping nonexistent file %s", path)
		}
		return
	}

	kb := float64(fi.Size()) / 1024
	if GlobalOpt.Verbose {
		log.Printf("Uploading %s (%.1fkb)", path, kb)
	}

	f, err := repoStore.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	_, err = apiclient.BuildData.Upload(fileSpec, f)
	if err != nil {
		log.Fatal(err)
	}

	if GlobalOpt.Verbose {
		log.Printf("Uploaded %s (%.1fkb)", path, kb)
	}
}
