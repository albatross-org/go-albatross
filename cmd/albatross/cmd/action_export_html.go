package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/albatross-org/go-albatross/entries"
	"github.com/plus3it/gorecurcopy"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

// ActionExportHTMLCmd represents the 'export html' command.
var ActionExportHTMLCmd = &cobra.Command{
	Use:     "html",
	Aliases: []string{"static"},
	Short:   "export the store as a HTML webpage",
	Long: `html exports the store a static HTML webpage.
	
For example, running

	$ albatross get -p school -o my-site/
	
Would format the store in HTML and output a fully working webpage in the given directory. If you want
to serve the site as well as outputting the site, you can use:

	$ albatross get -p school -o my-site/ --serve
	
Or, to serve the site only without creating a folder:

	$ albatross get -p school --serve-only

You can also use the --open and --open-path options to quickly open the served site in the browser.

	$ albatross get -p school --serve-only --open
	# Opens the webpage in the default system browser, after you quit the folder is removed

	$ albatross get -p school --serve-only --open-path "school/a-level/further-maths/topics/complex-numbers"
	# Opens the webpage for the entry 'school/a-level/further-maths/topics/complex-numbers' and removes the folder once you quit.
	
If you're looking to export the store as Hugo posts, you're probably looking for the 'albatross export hugo' command, though you
can output the Hugo folder used to build the site with this command:

	$ albatross get -p school -o my_hugo_site --hugo-dir

The main difference is 'export hugo' will generate the posts for an existing Hugo site, this creates a standalone site used for
quickly looking at a subset of the store.

Organisation
------------

You can access the HTML version of any entry by going to http://localhost:1313/posts/<path> where <path> is the path to the entry.

A homepage is also generated that contains 3 navigational aids:


	PAGE	        DESCRIPTION                                                          FLAG
	Info	        Small amount of information about the command used to generate       --page-info
			        the page and the amount of entries it matched.
			
	Tags            An alphabetic list of all tags, with links to pages containing       --page-tags
			        all entries with that tag.
			
	Paths           An alphabetic list of all paths to every entry that was matched.     --page-paths


	Chronological   A chronological list of all entries in the store.					 --page-chronological
	

You can disable any of these pages by setting its corresponding flag to false, i.e. --page-info=false.

Finally, you can also adjust the title of the webpage using the --site-title flag which defaults to "Albatross Export <DATE>".

	$ albatross get -p school export html --site-title "School Notes" --output notes-website/

Post Generation
---------------

A few options are provided that give a little bit of control over how the entries are generated.

--pre-insert will write some text before the entry content. This is useful for adding shortcodes:

	$ albatross get -p school/ export hugo -o content/posts/albatross --pre-insert "{{< toc >}}"
	# Entries now look like:
	# ---
	# ...metadata...
	# ---
	# {{< toc >}}
	# ...entry contents...

--post-insert does the same but writes it at the end of the entry instead.

	$ albatross get -p school/ export hugo -o content/posts/albatross --post-insert "{{< comments >}}"
	# Entries now look like:
	# ---
	# ...metadata...
	# ---
	# ...entry contents...
	# {{< comments >}}

--show-metadata appends a small section to each entry's page which shows the original metadata of each entry:

	$ albatross get -p school/ export hugo -o content/posts/albatross
	# Entries now look like:
	# ---
	# ...metadata...
	# ---
	#
	# ...entry contents...
	#
	# ###### Metadata
	#  title: 'Original Title'
	#  date: '2021-01-14 09:51'
	#  tags: ["@?tags"]

--show-attachments appends a small section to each entry's page with links to any attachments to the entry:

	$ albatross get -p school/ export hugo -o content/posts/albatross
	# Entries now look like:
	# ---
	# ...metadata...
	# ---
	#
	# ...entry contents...
	#
	# ###### Attachments
	# - image.jpg
	# - movie.mov

--metadata can be used to specify custom metadata that will be added to each Hugo post. For example:

	$ albatross get -p school/ export hugo -o content/posts/albatross --metadata toc=true --metadata draft=true
	# Entries now look like:
	# ---
	# ...metadata...
	# toc: true
	# draft: true
	# ---
	# ...entry contents...

Custom metadata given will not show up in the metadata section if --show-metadata is passed.	

Sub-Entries that Don't Match
----------------------------

Consider the case where you have one top-level entry you _want_ to match that contains a folder with attachments and also more
entries that you _don't_ want to match, like this:

For example:

  an-entry
    entry.md
    attachments/
      hello.jpg              < These are fine to include in the output.
      message.txt            <
    super-secret/            < We don't want to copy this folder!
      entry.md
    attachments-and-secret/
      some-attachment.jpg    < This is fine
      even-more-secret/      < This isn't!
		entry.md
		
We should be left with something that looks like this:

  an-entry/
    entry.md
    attachments/
      hello.jpg
      message.txt
    attachments-and-secret/
	  some-attachment.jpg    < This is fine.

These are the same rules that 'albatross export store' follow.

Skeletons
---------

A skeleton is how the command knows how to format the website and can be specified using the --skeleton command, as a
path to a folder or a URL to a Git repository. The default one is https://github.com/albatross/hugo-albatross-skeleton.

Skeletons differ from Hugo themes because they are much more static. The aim of a hugo theme is to allow the creation of
many different sites, skeletons basically just allow one. The program inserts a few folders and files into the directory
given and doesn't do any more configuration.

These are:

	content/
		posts/        - This is where the entries are placed. 
		homepage/     - This is where the information pages are placed, like the one listing all paths or tags.
			index.md 
			info.md            - Inserted page containing information about how the store was generating and information about entries.
			tags.md            - Page containing a table of tags and links to pages which display all tags.
			paths.md           - Page containing all paths in the store.
			chronological.md   - Page containing a chronological list of entries.
		
	config.toml     - This file is modified to include a correct title, by default "Albatross Export".
	
At the moment, it would probably be a lot of work to create a skeleton any different than the default one.`,

	Run: func(cmd *cobra.Command, args []string) {
		_, collection, list := getFromCommand(cmd)

		outputDir, err := cmd.Flags().GetString("output")
		checkArg(err)

		serve, err := cmd.Flags().GetBool("serve")
		checkArg(err)
		serveOnly, err := cmd.Flags().GetBool("serve-only")
		checkArg(err)
		port, err := cmd.Flags().GetString("port")
		checkArg(err)
		openBrowser, err := cmd.Flags().GetBool("open")
		checkArg(err)
		openPath, err := cmd.Flags().GetString("open-path")
		checkArg(err)

		if openPath != "" {
			openBrowser = true
		}

		outputHugoDir, err := cmd.Flags().GetBool("hugo-dir")
		checkArg(err)
		hugoDir := tempHugoDir()

		title, err := cmd.Flags().GetString("site-title")
		checkArg(err)
		if title == "" {
			title = "Albatross Export " + time.Now().Format("2006-01-02")
		}

		pageInfo, err := cmd.Flags().GetBool("page-info")
		checkArg(err)
		pageTags, err := cmd.Flags().GetBool("page-tags")
		checkArg(err)
		pagePaths, err := cmd.Flags().GetBool("page-paths")
		checkArg(err)
		pageChronological, err := cmd.Flags().GetBool("page-chronological")
		checkArg(err)

		preInsert, err := cmd.Flags().GetStringSlice("pre-insert")
		checkArg(err)
		postInsert, err := cmd.Flags().GetStringSlice("post-insert")
		checkArg(err)
		customMetadata, err := cmd.Flags().GetStringToString("metadata")
		checkArg(err)
		showMetadata, err := cmd.Flags().GetBool("show-metadata")
		checkArg(err)
		showAttachments, err := cmd.Flags().GetBool("show-attachments")
		checkArg(err)

		skeleton, err := cmd.Flags().GetString("skeleton")
		checkArg(err)

		if outputDir == "" && !serveOnly {
			fmt.Println("Please specify an output directory using the --output/-o flag")
			os.Exit(1)
		} else if outputDir != "" && serveOnly {
			fmt.Println("--serve-only flag given along with --output flag.")
			fmt.Println("--serve-only uses a temporary folder and has no output. Did you mean one or the other?")
			os.Exit(1)
		} else if outputDir == "" && serveOnly {
			// If we're serving the site and then deleting, we use the temporary
			// Hugo directory as the output location which will then be deleted later.
			outputDir = filepath.Join(hugoDir, "public")
		}

		err = checkOutputDest(outputDir)
		if err != nil && !serveOnly {
			// serveOnly means that we use the created temporary Hugo folder which hasn't been made properly yet.
			// Therefore we don't need to worry if it doesn't exist.
			fmt.Printf("Cannot output Hugo posts to %s:\n", outputDir)
			fmt.Println(err)
			os.Exit(1)
		}

		// Adjust any relative outputDir to an absolute path.
		// Mainly used so output is consistent.
		if !filepath.IsAbs(outputDir) {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Println("Could not establish current working directory:")
				fmt.Println(err)
				os.Exit(1)
			}

			outputDir = filepath.Join(cwd, outputDir)
		}

		copySkeleton(skeleton, hugoDir)

		fmt.Println("# Generating posts...")
		contentDir := filepath.Join(hugoDir, "content", "posts")
		generatePostsFolder(
			contentDir,
			list,
			collection,
			"posts",
			preInsert,
			postInsert,
			customMetadata,
			showMetadata,
			showAttachments,
		)

		homepageDir := filepath.Join(hugoDir, "content", "homepage")
		if pageInfo {
			fmt.Println("# Generating info.md page...")
			generateInfoPage(
				filepath.Join(homepageDir, "info.md"),
				"albatross "+strings.Join(os.Args[1:], " "),
				list,
				pageTags, pagePaths, pageChronological,
			)
		}

		if pageTags {
			fmt.Println("# Generating tags.md page...")
			generateTagsPage(
				filepath.Join(homepageDir, "tags.md"),
				list, collection,
			)
		}

		if pagePaths {
			fmt.Println("# Generating paths.md page...")
			generatePathsPage(
				filepath.Join(homepageDir, "paths.md"),
				list, collection,
			)
		}

		if pageChronological {
			fmt.Println("# Generating chronological.md page...")
			generateChronologicalPage(
				filepath.Join(homepageDir, "chronological.md"),
				list, collection,
			)
		}

		fmt.Println("# Updating config file...")
		updateConfigFile(
			filepath.Join(
				hugoDir,
				"config.toml",
			),
			title,
		)

		// If we're outputting just the Hugo directory only, we can stop here.
		if outputHugoDir {
			fmt.Println("# Moving outputted Hugo directory to", outputDir)
			os.Rename(hugoDir, outputDir)
			cleanupHugoDir(hugoDir)
			return
		}

		// Attempt to build the Hugo site.
		buildHugoSite(hugoDir, outputDir)

		if serve || serveOnly {
			serveHugoSite(outputDir, port, openBrowser, openPath)

		} else {
			fmt.Println("\n\n\n# Done! The HTML has been exported to", outputDir)
			cleanupHugoDir(hugoDir)
			return
		}

		fmt.Println("# Cleaning up...")
		fmt.Println("# Removing", hugoDir)
		cleanupHugoDir(hugoDir)

		if !serveOnly {
			fmt.Println("# Done! The HTML has been exported to", outputDir)
		} else {
			fmt.Println("# Done!")
		}

	},
}

// copySkeleton either copies a folder containing a skeleton site or clones it using Git if it's a URL.
func copySkeleton(skeleton, dest string) {
	if strings.HasPrefix(skeleton, "http") {
		// Is probably a Git URL.
		fmt.Printf("# Cloning git repository %s to %s\n", skeleton, dest)

		command := exec.Command("git", "clone", skeleton, dest)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		err := command.Run()
		if err != nil {
			fmt.Println("# Error cloning skeleton site git repository. There is likely additional output above.")
			fmt.Println(err)
		}

	} else if f, err := os.Stat(skeleton); !os.IsNotExist(err) && f.IsDir() {
		// Is a folder.
		fmt.Println("# Copying skeleton folder", skeleton)
		err := gorecurcopy.CopyDirectory(skeleton, dest)
		if err != nil {
			fmt.Println("Could not copy skeleton folder to destination:")
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Path/URL %q is not a valid URL or folder. Please try with a different argument.\n", skeleton)
		os.Exit(1)
	}
}

func init() {
	ActionExportCmd.AddCommand(ActionExportHTMLCmd)

	ActionExportHTMLCmd.Flags().StringP("output", "o", "", "output location of the static HTML")

	ActionExportHTMLCmd.Flags().Bool("serve", false, "serve the site")
	ActionExportHTMLCmd.Flags().Bool("serve-only", false, "serve the site and then delete folder")
	ActionExportHTMLCmd.Flags().String("port", "1313", "if the --serve option is given, use this port")

	ActionExportHTMLCmd.Flags().Bool("open", false, "if the --serve option is given, open a browser automatically")
	ActionExportHTMLCmd.Flags().String("open-path", "", "open the entry at this path rather than the homepage. It will set --open if given.")

	ActionExportHTMLCmd.Flags().Bool("hugo-dir", false, "instead of building the site, output the Hugo directory itself")

	ActionExportHTMLCmd.Flags().String("site-title", "", "the title of the generated site, default 'Albatross Export <DATE>'")
	ActionExportHTMLCmd.Flags().Bool("page-info", true, "output an info page. This is used as a homepage.")
	ActionExportHTMLCmd.Flags().Bool("page-tags", true, "output a page containing all tags")
	ActionExportHTMLCmd.Flags().Bool("page-paths", true, "output a page containing all paths")
	ActionExportHTMLCmd.Flags().Bool("page-chronological", true, "output a page containing all entries sorted chronologically")

	ActionExportHTMLCmd.Flags().String("skeleton", "https://github.com/albatross-org/hugo-albatross-skeleton", "path or URL to skeleton folder structure. See help for details.")

	ActionExportHTMLCmd.Flags().StringSlice("pre-insert", []string{}, "insert these strings before entry content, useful for inserting shortcodes")
	ActionExportHTMLCmd.Flags().StringSlice("post-insert", []string{}, "insert these strings after entry content, useful for inserting shortcodes")
	ActionExportHTMLCmd.Flags().StringToStringP("metadata", "m", map[string]string{}, "insert this metadata into every entry, such as '-m draft=true'")

	ActionExportHTMLCmd.Flags().Bool("show-metadata", false, "print a section at the end of every entry showing the original metadata")
	ActionExportHTMLCmd.Flags().Bool("show-attachments", false, "print a section at the end of every entry showing links to attachments")
}

// tempHugoDir creates and returns the path for a temporary folder used to build a site.
func tempHugoDir() (path string) {
	tmpDir, err := ioutil.TempDir("", "albatross-hugo-build")
	if err != nil {
		fmt.Println("Could not create temporary directory:")
		fmt.Println(err)
		os.Exit(1)
	}

	// Append 'site' onto the path as this is where the site will actually be created.
	return filepath.Join(tmpDir, "site")
}

// cleanupHugoDir deletes a temporary Hugo directory.
func cleanupHugoDir(path string) {
	if !strings.Contains(path, "/tmp/albatross-hugo-build") || filepath.Base(path) != "site" {
		fmt.Println("Temporary directory does not start with /tmp/albatross-hugo-build/.../site, refusing to delete.")
		os.Exit(1)
	}

	// Remove the 'site' bit on the end, we're deleting everything.
	dir := filepath.Dir(path)
	err := os.RemoveAll(dir)
	if err != nil {
		fmt.Printf("Couldn't remove directory %s:", dir)
		fmt.Println(err)
		os.Exit(1)
	}
}

// generateInfoPage creates an info.md page at the correct location.
func generateInfoPage(dest string, command string, list entries.List, pageTags, pagePaths, pageChronological bool) {
	template := `+++
title = "Info"
weight = 1

[widget]
	handler = "blank"
+++
	
# Albatross Export
This website was generated %s using the following command:

%s

Which matched %s entries.`

	if pageTags || pagePaths || pageChronological {
		template += "\n\n## Jump To"
	}

	if pageTags {
		template += "\n- [Tags Search](#tags)"
	}

	if pagePaths {
		template += "\n- [Paths Search](#paths)"
	}

	if pageChronological {
		template += "\n- [Chronological Search](#chronological)"
	}

	date := time.Now().Format("`2006-01-02 15:04`")
	command = "```\n" + command + "\n```"
	entriesCount := fmt.Sprintf("`%d`", len(list.Slice()))

	output := fmt.Sprintf(template, date, command, entriesCount)
	err := ioutil.WriteFile(dest, []byte(output), 0644)
	if err != nil {
		fmt.Printf("Error writing info.md file to %s:\n", dest)
		fmt.Println(err)
		os.Exit(1)
	}
}

// generateTagsPage creates an tags.md page at the correct location.
// BUG: old/unused tag syntax for '@!public' is grouped into the same page as '@?public' by Hugo.
func generateTagsPage(dest string, list entries.List, collection *entries.Collection) {
	template := `+++
title = "Tags"
weight = 2

[widget]
	handler = "blank"
+++

## Tags
|Tag|Entries|
|-|-|
`

	var out bytes.Buffer
	tagMap := make(map[string]int)

	for _, entry := range list.Slice() {
		for _, tag := range entry.Tags {
			tagMap[tag]++
		}
	}

	tagList := []string{}
	for tag := range tagMap {
		tagList = append(tagList, tag)
	}

	sort.Strings(tagList)

	for _, tag := range tagList {
		out.WriteString("|[`")
		out.WriteString(tag)
		out.WriteString("`](/tags/")
		out.WriteString(strings.TrimLeft(tag, "@?!"))
		out.WriteString(")|")
		out.WriteString(fmt.Sprint(tagMap[tag]))
		out.WriteString("|\n")
	}

	output := template + out.String()
	err := ioutil.WriteFile(dest, []byte(output), 0644)
	if err != nil {
		fmt.Printf("Error writing tags.md file to %s:\n", dest)
		fmt.Println(err)
		os.Exit(1)
	}
}

// generatePathsPage creates an paths.md page at the correct location.
// TODO: have this generate a tree like the `ls` command does.
func generatePathsPage(dest string, list entries.List, collection *entries.Collection) {
	template := `+++
title = "Paths"
weight = 3

[widget]
	handler = "blank"
+++

## Paths
`

	sorted := list.Sort(entries.SortPath)
	var out bytes.Buffer

	for _, entry := range sorted.Slice() {
		out.WriteString("- [`")
		out.WriteString(entry.Path)
		out.WriteString("`](/posts/")
		out.WriteString(entry.Path)
		out.WriteString(")\n")
	}

	output := template + out.String()
	err := ioutil.WriteFile(dest, []byte(output), 0644)
	if err != nil {
		fmt.Printf("Error writing paths.md file to %s:\n", dest)
		fmt.Println(err)
		os.Exit(1)
	}
}

// generateChronologicalPage creates an tags.md page at the correct location.
func generateChronologicalPage(dest string, list entries.List, collection *entries.Collection) {
	template := `+++
title = "Chronological"
weight = 4

[widget]
	handler = "blank"
+++

## Chronological
`

	var out bytes.Buffer
	var currentYear, currentMonth string

	sorted := list.Sort(entries.SortDate)
	for _, entry := range sorted.Slice() {
		if currentYear != entry.Date.Format("2006") {
			currentYear = entry.Date.Format("2006")
			out.WriteString("### ")
			out.WriteString(currentYear)
			out.WriteString("\n")
		}

		if currentMonth != entry.Date.Format("January") {
			currentMonth = entry.Date.Format("January")
			out.WriteString("#### ")
			out.WriteString(currentMonth)
			out.WriteString("\n")
			out.WriteString("\n|Title|Date|\n|-|-|\n")
		}

		out.WriteString("|[")
		out.WriteString(entry.Title)
		out.WriteString("](/posts/")
		out.WriteString(entry.Path)
		out.WriteString(")|`")
		out.WriteString(entry.Date.Format("2006-01-02 15:04"))
		out.WriteString("`|\n")
	}

	output := template + out.String()
	err := ioutil.WriteFile(dest, []byte(output), 0644)
	if err != nil {
		fmt.Printf("# Error writing chronological.md file to %s:\n", dest)
		fmt.Println(err)
		os.Exit(1)
	}
}

// updateConfigFile updates a configuration file to make it use the correct information.
func updateConfigFile(dest string, title string) {
	configBytes, err := ioutil.ReadFile(dest)
	if err != nil {
		fmt.Println("# Couldn't read existing config file. Are you sure there's a config.toml in the scaffold?")
		fmt.Println(err)
		os.Exit(1)
	}

	config := string(configBytes)
	newConfig := `title = "` + title + `"` + "\n" + config

	err = ioutil.WriteFile(dest, []byte(newConfig), 0644)
	if err != nil {
		fmt.Println("# Error updating config in skaffold.")
		fmt.Println(err)
		os.Exit(1)
	}
}

// gracefulShutdown will gracefully stop a HTTP server.
// Courtesy of https://marcofranssen.nl/go-webserver-with-graceful-shutdown/
func gracefulShutdown(server *http.Server, quit <-chan os.Signal, done chan<- bool) {
	<-quit
	fmt.Println("\n# Shutting down webserver...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("# Couldn't gracefully shut down the webserver:")
		fmt.Println("# Error:")
		fmt.Println(err)
		os.Exit(1)
	}
	close(done)
}

// buildHugoSite attempts to use the Hugo command to build the site.
func buildHugoSite(hugoDir, outputDir string) {
	fmt.Println("# Building Hugo site at", hugoDir)
	fmt.Println("")
	command := exec.Command("hugo", "--source", hugoDir, "--destination", outputDir)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		fmt.Println("\n# Error building Hugo site. There is likely additional information above.")
		fmt.Println("# Error:", err)
		os.Exit(1)
	}
}

func serveHugoSite(outputDir, port string, openBrowser bool, openPath string) {
	fs := http.FileServer(http.Dir(outputDir))
	router := http.NewServeMux()
	router.Handle("/", fs)

	server := &http.Server{
		Addr:         fmt.Sprint(":", port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// We don't want to quit the whole program on a CTRL-C, so here we capture interrupts and use them to shut down the server.
	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go gracefulShutdown(server, quit, done)

	fmt.Printf("\n# Starting HTTP server on localhost:%s ...\n", port)
	fmt.Println("# Press CTRL-C to stop the server.")

	// We have to start listening ourselves so that the browser can connect properly.
	// See https://stackoverflow.com/questions/32738188/how-can-i-start-the-browser-after-the-server-started-listening
	l, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		fmt.Printf("# Error couldn't listen on port %s:\n", port)
		fmt.Println(err)
		os.Exit(1)
	}

	if openBrowser {
		url := ""

		if openPath != "" {
			url = fmt.Sprintf("http://localhost:%s/posts/%s", port, openPath)
		} else {
			url = fmt.Sprintf("http://localhost:%s", port)
		}
		err = open.Start(url)

		if err != nil {
			fmt.Println("# Error opening browser:")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	err = server.Serve(l)
	if err != nil && err != http.ErrServerClosed {
		fmt.Printf("# Error: Couldn't listen on port %s:\n", port)
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("# Stopped HTTP server.")
}
