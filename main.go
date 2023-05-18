package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var MyHostName string
var FilterStr string
var Title string

func init() {
	MyHostName, _ = os.Hostname()
}
func main() {
	var host string
	var port int
	flag.StringVar(&host, "m", MyHostName, "redis server hostname")
	flag.IntVar(&port, "p", 3978, "redis server port")
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", host, port)
	Title = "Redis(" + addr + ")"
	// 创建tview应用程序
	app := new(App).init()
	app.SetTitle(Title)
	app.ConnectRedis(addr)

	// 获取所有的key
	keys, err := app.client.GetAllKeys()
	if err != nil {
		log.Fatalf("Failed to fetch Redis keys: %s", err)
	}
	sort.Strings(keys)
	app.SetListKey(keys)

	app.list.SetInputCapture(app.handleListKeyEvent)
	app.table.SetInputCapture(app.handleTableKeyEvent)

	// 启动应用程序
	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run application: %s", err)
	}
}

type App struct {
	client   *RedisClient
	app      *tview.Application
	prompt   *tview.TextView
	list     *tview.List
	table    *tview.Table
	pages    *tview.Pages
	viewpage *tview.Flex
	showPage *tview.TextView
	editPage *tview.TextArea
}

func (a *App) init() *App {
	app := tview.NewApplication()
	prompt := tview.NewTextView()

	list := tview.NewList()
	list.SetSecondaryTextColor(tcell.ColorRed)
	list.SetSelectedTextColor(tcell.ColorNames["black"])
	list.SetSelectedFocusOnly(true)
	flex := tview.NewFlex()
	flex.AddItem(list, 0, 1, true)
	flex.SetDirection(tview.FlexRow)

	table := tview.NewTable()

	pages := tview.NewPages()
	a.showPage = tview.NewTextView().SetWordWrap(true)
	a.showPage.SetDisabled(true)
	a.showPage.SetScrollable(true)
	a.editPage = tview.NewTextArea().SetWordWrap(true)
	a.editPage.SetLabel("New:")

	viewFlex := tview.NewFlex()
	viewFlex.SetDirection(tview.FlexRow)
	viewFlex.SetBorder(true)

	Homepage := tview.NewGrid().
		SetRows(1, 0).
		SetColumns(20, 0).
		SetBorders(true).
		AddItem(prompt, 0, 0, 1, 2, 0, 0, false).
		AddItem(flex, 1, 0, 1, 1, 0, 0, false).
		AddItem(table, 1, 1, 1, 1, 0, 0, true)

	pages.AddPage("Homepage", Homepage, true, true)
	pages.AddPage("Viewpage", viewFlex, true, false)

	app.SetRoot(pages, true)
	app.SetFocus(flex)

	a.app = app
	a.prompt = prompt
	a.table = table
	a.viewpage = viewFlex
	a.pages = pages
	a.list = list
	return a
}

func (a *App) ConnectRedis(addr string) {
	client := newRedisClient(addr, "", 0)
	// 测试Redis连接
	_, err := client.TestPing()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %s", err)
	}
	a.client = client
}
func (a *App) SetTitle(title ...string) {
	a.prompt.SetText(strings.Join(title, " --> "))
}

func (a *App) SetListKey(keys []string) {
	a.list.Clear()
	for _, key := range keys {
		t, _ := a.client.GetType(key)
		a.list.AddItem(key, t, 0, nil)
	}
}
func (a *App) SetTableHash(hash map[string]string) {
	a.table.Clear()
	a.table.SetCell(0, 0, tview.NewTableCell("Field").SetSelectable(false))
	a.table.SetCell(0, 1, tview.NewTableCell("Value").SetSelectable(false))
	for field, val := range hash {
		a.table.SetCellSimple(a.table.GetRowCount(), 0, field)
		a.table.SetCellSimple(a.table.GetRowCount()-1, 1, val)
	}
}
func (a *App) SetTableValues(values []interface{}) {
	a.table.Clear()
	for i, val := range values {
		a.table.SetCellSimple(i, 0, val.(string))
	}
}

func (a *App) SetTableFocus() {
	a.app.SetFocus(a.table)
	a.table.SetSelectable(true, false)
	selectedIndex := a.list.GetCurrentItem()
	_, t := a.list.GetItemText(selectedIndex)
	if t == "hash" {
		a.table.Select(1, 0)
	} else {
		a.table.Select(0, 0)
	}
}
func (a *App) handleListKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyF5:
		a.table.Clear()
		keys, err := a.client.GetAllKeys()
		if err != nil {
			log.Fatalf("Failed to fetch Redis keys: %s", err)
		}
		sort.Strings(keys)
		a.SetListKey(keys)
	case tcell.KeyDown, tcell.KeyUp:
		offset := -1
		if event.Key() == tcell.KeyDown {
			offset = 1
		}

		selectedIndex := a.list.GetCurrentItem()
		index := selectedIndex + offset
		if index < 0 {
			index = a.list.GetItemCount() + index
		}
		if index >= a.list.GetItemCount() {
			index = 0
		}

		selectedKey, t := a.list.GetItemText(index)
		a.SetTitle(Title, selectedKey)
		a.UpdateListValue(selectedKey, t)
		FilterStr = ""
	case tcell.KeyRight:
		a.SetTableFocus()
		selectedIndex := a.list.GetCurrentItem()
		selectedKey, _ := a.list.GetItemText(selectedIndex)
		a.SetTitle(Title, selectedKey)
		FilterStr = ""
		return nil
	case tcell.KeyLeft:
		return nil
	case tcell.KeyEsc:
		a.SetTitle(Title)
		FilterStr = ""
	case tcell.KeyRune:
		FilterStr += string(event.Rune())
		a.SetTitle(Title, FilterStr)
		for i := 0; i < a.list.GetItemCount(); i++ {
			key, t := a.list.GetItemText(i)
			if strings.Contains(strings.ToUpper(key), strings.ToUpper(FilterStr)) {
				a.list.SetCurrentItem(i)
				a.UpdateListValue(key, t)
				break
			}
		}

	}

	return event
}

func (a *App) handleTableKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	selectedIndex := a.list.GetCurrentItem()
	selectedKey, t := a.list.GetItemText(selectedIndex)
	switch event.Key() {
	case tcell.KeyF2:
		selectedRow, selectedColumn := a.table.GetSelection()
		if t == "hash" {
			selectedColumn = 1
		}
		selectedField := a.table.GetCell(selectedRow, 0).Text
		selectedText := a.table.GetCell(selectedRow, selectedColumn).Text
		a.SetTitle(Title, fmt.Sprintf("%s ==> (%s)", selectedKey, selectedField))
		a.showPage.SetText(selectedText)
		a.SetViewPage(true, false)

		a.editPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				// Grab the text from the input field
				newValue := a.editPage.GetText()
				newValue = strings.Trim(newValue, "\n")
				newValue = strings.TrimSpace(newValue)
				if selectedColumn == 1 {
					err := a.client.SetHashKeyVal(selectedKey, selectedField, newValue)
					if err != nil {
						log.Fatalf("Failed to update Redis value: %s", err)
					}
					a.table.SetCellSimple(selectedRow, 1, newValue)
				} else {
					err := a.client.SetStringKeyVal(selectedKey, newValue)
					if err != nil {
						log.Fatalf("Failed to update Redis value: %s", err)
					}
					a.table.SetCellSimple(selectedRow, 0, newValue)
				}
				a.SetViewPage(false, false)
				return nil
			case tcell.KeyEsc:
				a.SetViewPage(false, false)
			}
			return event
		})

	case tcell.KeyLeft:
		a.app.SetFocus(a.list)
		a.table.SetSelectable(false, false)
		a.SetTitle(Title, selectedKey)
		FilterStr = ""
		return nil
	case tcell.KeyRight:
		return nil
	case tcell.KeyEnter:
		selectedRow, selectedColumn := a.table.GetSelection()
		if t == "hash" {
			selectedColumn = 1
		}
		selectedText := a.table.GetCell(selectedRow, selectedColumn).Text
		a.showPage.SetText(selectedText)
		a.showPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEnter {
				a.SetViewPage(false, false)
			}
			return event
		})
		a.SetViewPage(true, true)

	case tcell.KeyEsc, tcell.KeyDown, tcell.KeyUp:
		a.SetTitle(Title, selectedKey)
		FilterStr = ""
	case tcell.KeyRune:
		if a.table.GetRowCount() < 2 {
			break
		}
		selectedIndex := a.list.GetCurrentItem()
		listKey, t := a.list.GetItemText(selectedIndex)
		FilterStr += string(event.Rune())
		a.SetTitle(Title, listKey, FilterStr)
		j := 0
		if t == "hash" {
			j = 1
		}
		for i := j; i < a.table.GetRowCount(); i++ {
			c := a.table.GetCell(i, 0)
			if strings.Contains(strings.ToUpper(c.Text), strings.ToUpper(FilterStr)) {
				a.table.Select(i, 0)
				break
			}
		}
	}

	return event
}
func (a *App) SetViewPage(show bool, viewOnly bool) {
	if show {
		a.viewpage.Clear()
		if !viewOnly {
			a.viewpage.AddItem(a.showPage, 0, 1, false)
			a.showPage.SetLabel("Old:")
			a.viewpage.AddItem(a.editPage, 0, 1, true)
			a.app.SetFocus(a.editPage)
		} else {
			a.viewpage.AddItem(a.showPage, 0, 1, true)
			a.showPage.ScrollToBeginning()
		}
		a.pages.ShowPage("Viewpage")
	} else {
		a.editPage.SetText("", false)
		a.showPage.SetLabel("")
		a.editPage.SetText("", false)
		a.pages.HidePage("Viewpage")
		a.viewpage.Clear()
		a.SetTableFocus()
	}
}

func (a *App) UpdateListValue(selectedKey, t string) {
	if t == "hash" {
		hash, err := a.client.GetAllHashValues(selectedKey)
		if err != nil {
			log.Fatalf("Failed to fetch Redis hash: %s", err)
		}
		a.SetTableHash(hash)
	} else {
		// 获取选中key的值
		values, err := a.client.GetValues([]string{selectedKey})
		if err != nil {
			log.Fatalf("Failed to fetch Redis<%s> value: %s", selectedKey, err)
		}
		a.SetTableValues(values)
	}
}

func (a *App) Run() error {
	return a.app.Run()
}
