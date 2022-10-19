package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	_ "github.com/mattn/go-sqlite3"
)

func showAuthors(db *sql.DB) error {
	rows, err := db.Query(`
        SELECT
            a.author_id,
            a.author
        FROM
            authors a
        ORDER BY
            CAST(a.author_id AS INTEGER)
    `)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var authorID, author string
		err = rows.Scan(&authorID, &author)
		if err != nil {
			return err
		}
		fmt.Printf("%s %s\n", authorID, author)
	}
	return nil
}

func showTitles(db *sql.DB, authorID string) error {
	rows, err := db.Query(`
        SELECT
            c.title_id,
            c.title
        FROM
            contents c
        WHERE
            c.author_id = ?
        ORDER BY
            CAST(c.title_id AS INTEGER)
    `, authorID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var titleID, title string
		err = rows.Scan(&titleID, &title)
		if err != nil {
			return err
		}
		fmt.Printf("% 5s %s\n", titleID, title)
	}
	return nil
}

func showContent(db *sql.DB, authorID string, titleID string) error {
	var content string
	err := db.QueryRow(`
        SELECT
            c.content
        FROM
            contents c
        WHERE
            c.author_id = ?
        AND c.title_id = ?
    `, authorID, titleID).Scan(&content)
	if err != nil {
		return err
	}
	fmt.Println(content)
	return nil
}

func queryContent(db *sql.DB, query string) error {
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return err
	}

	seg := t.Wakati(query)
	rows, err := db.Query(`
        SELECT
            a.author_id,
            a.author,
            c.title_id,
            c.title
        FROM
            contents c
        INNER JOIN authors a
            ON a.author_id = c.author_id
        INNER JOIN contents_fts f
            ON c.rowid = f.docid
            AND words MATCH ?
    `, strings.Join(seg, " "))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var authorID, author string
		var titleID, title string
		err = rows.Scan(&authorID, &author, &titleID, &title)
		if err != nil {
			return err
		}
		fmt.Printf("%s % 5s: %s (%s)\n", authorID, titleID, title, author)
	}
	return nil
}

const usage = `
Usage of ./aozora-search [sub-command] [...]:
  -d string
        database (default "database.sqlite")

Sub-commands:
    authors
    titles  [AuthorID]
    content [AuthorID] [TitleID]
    query   [Query]
`

func main() {
	var dsn string
	flag.StringVar(&dsn, "d", "database.sqlite", "database")
	flag.Usage = func() {
		fmt.Print(usage)
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	switch flag.Arg(0) {
	case "authors":
		err = showAuthors(db)
	case "titles":
		if flag.NArg() != 2 {
			flag.Usage()
			os.Exit(2)
		}
		err = showTitles(db, flag.Arg(1))
	case "content":
		if flag.NArg() != 3 {
			flag.Usage()
			os.Exit(2)
		}
		err = showContent(db, flag.Arg(1), flag.Arg(2))
	case "query":
		if flag.NArg() != 2 {
			flag.Usage()
			os.Exit(2)
		}
		err = queryContent(db, flag.Arg(1))
	}

	if err != nil {
		log.Fatal(err)
	}
}
