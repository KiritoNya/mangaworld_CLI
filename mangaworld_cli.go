package main

import (
	"errors"
	"fmt"
	"github.com/KiritoNya/mangaworld"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type downloadInput struct {
	name      string
	chapter int
	volume  int
}

func main() {
	var di downloadInput

	app := &cli.App{
		Name:     "mangaworld",
		Version:  "v1.0.0",
		Compiled: time.Now(),
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "KiritoNya",
				Email: "watashiwayuridaisuki@gmail.com",
			}},
		Copyright: "(c) 1999 Serious Enterprise",
		HelpName:  "mangaworld",
		Usage:     "a simple mangaworld tool",
		//UsageText: "nhentai - a simple nhentai downloader",
		Commands: []*cli.Command{
			&cli.Command{
				Name:            "download",
				Aliases:         []string{"dw"},
				Category:        "Download",
				Usage:           "do the download of manga",
				UsageText:       "download [argument]",
				Description:     "Download item from mangaworld",
				ArgsUsage:       "[arrgh]",
				SkipFlagParsing: false,
				HideHelp:        false,
				Hidden:          false,
				HelpName:        "download",
				Action: func(c *cli.Context) error {

					//Get arguments manga name
					di.name = c.Args().Get(0)
					if c.NArg() < 1 {
						return errors.New("<Error> Name manga arguments not found, view the \"--help\" command")
					}
					if c.NArg() > 1 {
						return errors.New("<Error> Too many arguments, view the \"--help\" command")
					}

					//Search the manga
					manga, err := search(di.name)
					if err != nil {
						return err
					}

					//Get chapters
					manga.GetChaptersNum()
					manga.GetChapters(1, manga.ChaptersNum)
					//Choose chapters
					err = chapterChosen(&manga)
					if err != nil {
						return err
					}

					err = downloadChapter(&manga)
					if err != nil {
						return err
					}

					return nil
				},
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "chapter", Value: 0, Aliases: []string{"ch"}, Usage: "Use this flag for specif to download a chapter", Destination: &di.chapter},
					&cli.IntFlag{Name: "volume", Value: 0, Aliases: []string{"vol"}, Usage: "Use this flag for specif to download a volume", Destination: &di.volume},
				},
				OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
					fmt.Fprintf(c.App.Writer, "for shame\n")
					return err
				},
			},
		},
		/*Flags: []cli.Flag{
			&cli.StringFlag{Name: "input", Value: "", Aliases: []string{"i"}, Usage: "path of the file containing the links", Destination: &in.InFile},
			&cli.BoolFlag{Name: "id", Value: false, Usage: "use this flag to insert only the item ID", Destination: &in.IdMode},
			&cli.BoolFlag{Name: "not-json", Aliases: []string{"njs"}, Value: false, Usage: "Use this flag to not write the item info in a JSON file (required if don't have \"config.yml\")", Destination: &in.JsonMode},
			&cli.StringFlag{Name: "output", Value: "", Aliases: []string{"o"}, Usage: "destination path", Destination: &in.OutFile},
		},*/
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func search(name string) (mangaworld.Manga, error) {
	listManga := mangaworld.NewListManga()

	err := listManga.SearchByName(name)
	if err != nil {
		return mangaworld.Manga{}, errors.New("<Error> Manga doesn't exist")
	}

	listManga.AddTitles()

	prompt := promptui.Select{
		Label: "Seleziona il manga che vuoi leggere",
		Items: listManga.GetTitles(),
	}

	numChosen, _, err := prompt.Run()
	if err != nil {
		return mangaworld.Manga{}, err
	}

	return listManga.Mangas[numChosen], nil
}

func chapterChosen(manga *mangaworld.Manga) error {

	var start int
	var end int

	//Choose modality
	prompt := promptui.Select{
		Label: "Il manga da te scelto ha" + strconv.Itoa(manga.ChaptersNum) + "cosa vuoi fare?",
		Items: []string{"Scaricali tutti", "Scarica più capitoli", "Scarica un capitolo"},
	}

	modChosen, _, err := prompt.Run()
	if err != nil {
		return err
	}

	switch modChosen {

	//all chapters
	case 0:
		return nil

	//range chapters
	case 1:
		validate := func(input string) error {
			if !strings.Contains(input, "-") {
				return err
			} else {
				matrix := strings.Split(input, "-")

				num, err := strconv.Atoi(matrix[0])
				if err != nil || num < 1 || num > manga.ChaptersNum {
					return errors.New("Invalid number start")
				}

				num, err = strconv.Atoi(matrix[1])
				if err != nil || num < 1 || num > manga.ChaptersNum {
					return errors.New("Invalid number end")
				}
			}
			return nil
		}

		prompt := promptui.Prompt{
			Label:    "Range (EX: \"1-10\")",
			Validate: validate,
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}

		matrix := strings.Split(result, "-")
		start, _ = strconv.Atoi(matrix[0])
		end, _ = strconv.Atoi(matrix[1])

		manga.Chapters = manga.Chapters[start-1:end]
		return  nil

	//one chapter
	case 2:
		validate := func(input string) error {
			num, err := strconv.Atoi(input)
			if err != nil {
				return errors.New("Invalid number")
			}
			if num > manga.ChaptersNum || num < 0 {
				return errors.New("Invalid number, chapter doesn't exist")
			}
			return nil
		}

		prompt := promptui.Prompt{
			Label:    "Number",
			Validate: validate,
		}

		result, err := prompt.Run()
		if err != nil {
			return err
		}
		start, err := strconv.Atoi(result)
		manga.Chapters = manga.Chapters[:start]
		return nil
	}
	return nil
}

func downloadChapter(manga *mangaworld.Manga) error {

	fmt.Println(manga.Chapters)

	for _, chapter := range manga.Chapters {

		//Get number of chapter
		err := chapter.GetNumber()
		if err != nil {
			return err
		}

		//Create path
		path, err := createPath("C:\\Users\\KiritoNya\\Desktop\\Nuova", manga.Title, chapter.Number, manga.ChaptersNum)
		if err != nil {
			return err
		}
		fmt.Println(path)

		//Download all chapter pages
		err = chapter.Download(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func createPath(destPath string, mangaName string, chapterNum int, chaptersNum int ) (string, error) {
	finalDest := destPath + string(os.PathSeparator) + mangaName + string(os.PathSeparator) + createChapterNum("Ch", chapterNum, chaptersNum) + string(os.PathSeparator)
	err := os.MkdirAll(finalDest, 0700)
	if err != nil {
		return "", err
	}
	return finalDest, nil
}

func createChapterNum(prefix string, chapterNum int, chaptersNum int) string {
	str := prefix + " "
	if chapterNum < 10 {
		str += "0"
	}
	if chaptersNum > 99 && chapterNum < 100 {
		str += "0"
	}
	return str + strconv.Itoa(chapterNum)
}