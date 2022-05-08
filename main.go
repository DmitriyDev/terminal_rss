package main

import (
	"context"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jaytaylor/html2text"
	"github.com/mmcdole/gofeed"
	"github.com/rivo/tview"
)

const logo = `.__  __. __.  .__         .      
[__)(__ (__   [__) _  _. _| _ ._.
|  \.__).__)  |  \(/,(_](_](/,[  
`

type Source struct {
	Name   string
	Url    string
	Topics map[int]Topic
}

type Topic struct {
	Name   string
	Author string
	Text   string
}

var sources = map[int]Source{
	0: Source{
		Name:   "Jobs DOU PHP",
		Url:    "https://jobs.dou.ua/vacancies/feeds/?category=PHP",
		Topics: map[int]Topic{},
	},
	1: Source{
		Name:   "Jobs DOU Golang",
		Url:    "https://jobs.dou.ua/vacancies/feeds/?category=Golang",
		Topics: map[int]Topic{},
	},
	2: Source{
		Name:   "Jobs DOU Architect",
		Url:    "https://jobs.dou.ua/vacancies/feeds/?category=Architect",
		Topics: map[int]Topic{},
	},
}

func readRss(si int) {
	s := sources[si]
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(s.Url, ctx)
	if err != nil {
		panic(err.Error())
	}

	for i, item := range feed.Items {
		tp := Topic{
			Name:   "",
			Author: "",
			Text:   "",
		}
		if item.Author != nil {
			tp.Author = item.Author.Name
		}
		if item.Description != "" {
			tp.Text = item.Description
		}
		if item.Content != "" {
			tp.Text = item.Content
		}

		if item.Title != "" {
			tp.Name = item.Title
		}

		sources[si].Topics[i] = tp
	}

}

func showSources(blk *tview.TextView, t chan Source) {
	go func() {
		s := <-t
		blk.SetText(s.Name)

	}()
}

func showTopic(ra *RssApp) {
	go func() {
		s := <-ra.TopicChan
		// readRss(&s)
		createTopicTable(ra, s)
	}()
}

func createTopicTable(ra *RssApp, s Source) {

	ra.TopicBlock.Clear()

	for i, t := range s.Topics {
		ra.TopicBlock.SetCell(i, 0,
			tview.NewTableCell(t.Name).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignLeft))
	}

	ra.TopicBlock.Select(0, 0).SetFixed(1, 0).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			ra.Application.Stop()
		}
		if key == tcell.KeyEnter {
			ra.TopicBlock.SetSelectable(true, true)
		}
	}).SetSelectedFunc(func(row int, column int) {

		for i := range s.Topics {
			ra.TopicBlock.GetCell(i, 0).SetTextColor(tcell.ColorYellow).SetText(s.Topics[i].Name)
		}
		ra.ContentChan <- s.Topics[row]
		showContent(ra)

		ra.TopicBlock.GetCell(row, column).SetTextColor(tcell.ColorRed).SetText(s.Topics[row].Name)

	})
}

func showContent(ra *RssApp) {
	go func() {
		s := <-ra.ContentChan

		text, err := html2text.FromString(s.Text)
		if err != nil {
			panic(err)
		}

		ra.ContentBlock.SetText(s.Name + " \n\n" + text)
	}()
}

func createTable(ra *RssApp) {
	for i, sr := range sources {
		ra.SourcesBlock.SetCell(i, 0,
			tview.NewTableCell(sr.Name).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignLeft))
	}

	ra.SourcesBlock.Select(0, 0).SetFixed(0, 0).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			ra.Application.Stop()
		}
		if key == tcell.KeyEnter {
			ra.SourcesBlock.SetSelectable(true, true)
		}
	}).SetSelectedFunc(func(row int, column int) {

		for i := range sources {
			ra.SourcesBlock.GetCell(i, 0).SetTextColor(tcell.ColorYellow).SetText(sources[i].Name)
		}
		ra.TopicChan <- sources[row]
		showTopic(ra)

		ra.SourcesBlock.GetCell(row, column).SetTextColor(tcell.ColorRed).SetText(sources[row].Name)
	})

	ra.SourcesBlock.GetCell(0, 0).SetSelectable(true)
}

type RssApp struct {
	Application *tview.Application

	SourcesBlock *tview.Table

	TopicBlock *tview.Table
	TopicChan  chan Source

	ContentBlock *tview.TextView
	ContentChan  chan Topic

	Grid *tview.Grid
}

func (ra *RssApp) GetHeader() *tview.TextView {

	return tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetText(logo)

}

func (ra *RssApp) GetBody() *tview.Grid {
	grid := tview.NewGrid().
		SetRows(3, 0, 1).
		SetColumns(30, 40, 0).
		SetBorders(true).
		AddItem(ra.GetHeader(), 0, 0, 1, 3, 0, 0, false).
		AddItem(ra.GetFooter(), 2, 0, 1, 3, 0, 0, false)

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	grid.AddItem(ra.SourcesBlock, 0, 0, 0, 0, 0, 0, true).
		AddItem(ra.TopicBlock, 1, 0, 1, 3, 0, 0, false).
		AddItem(ra.ContentBlock, 0, 0, 0, 0, 0, 0, false)

	// Layout for screens wider than 100 cells.
	grid.AddItem(ra.SourcesBlock, 1, 0, 1, 1, 0, 100, true).
		AddItem(ra.TopicBlock, 1, 1, 1, 1, 0, 100, false).
		AddItem(ra.ContentBlock, 1, 2, 1, 1, 0, 100, false)

	return grid

}

func (ra *RssApp) GetFooter() *tview.TextView {
	return tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("⁀⊙﹏☉⁀ Ver 0.0.1 alpha")
}

func (ra *RssApp) Render() {
	ra.Grid = ra.GetBody()

	if err := ra.Application.SetRoot(ra.Grid, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func main() {

	for i := range sources {
		go readRss(i)
	}

	rssApp := RssApp{
		Application: tview.NewApplication(),

		SourcesBlock: tview.NewTable().SetBorders(false),

		TopicBlock: tview.NewTable().SetBorders(false),
		TopicChan:  make(chan Source),

		ContentBlock: tview.NewTextView().SetText("Content"),
		ContentChan:  make(chan Topic),
	}

	createTable(&rssApp)
	showTopic(&rssApp)
	showContent(&rssApp)

	rssApp.Render()

}
