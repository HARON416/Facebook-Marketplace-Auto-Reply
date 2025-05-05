package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"

	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
)

func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func main() {
	// Set up logging
	log.SetLevel(log.InfoLevel)
	log.Info("Starting Facebook Messenger automation")
	fmt.Println()

	// Request user input for reply message
	var replyMessage string
	fmt.Print("Enter your reply message: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		replyMessage = scanner.Text()
	}

	if replyMessage == "" {
		log.Error("Reply message is required")
		os.Exit(1)
	} else {
		log.Info("Reply message set", "message", replyMessage)
	}

	url := "https://web.facebook.com/messages"

	l := launcher.New().
		Leakless(true).
		NoSandbox(true).
		Headless(false)

	u := l.MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect().NoDefaultDevice()
	defer browser.MustClose()

	page := browser.MustPage(url).MustWaitLoad().MustWindowMaximize()

	for {
		if page.MustHas(`button[name="login"]`) {
			log.Info("LOG IN TO YOUR FACEBOOK ACCOUNT")
			time.Sleep(1 * time.Second)
			continue
		}

		page = page.MustWaitLoad().MustWaitDOMStable()
		break
	}

	log.Info("YOU ARE LOGGED IN")

	for {
		pageHasDialog, dialog, err := page.Has(`div[role="dialog"]`)
		if err != nil {
			log.Error("Error checking for dialog", "error", err)
			os.Exit(1)
		}

		if pageHasDialog {
			dialogIsRestoreChatHistory := strings.Contains(dialog.MustText(), "Enter your PIN to restore your chat history")

			if dialogIsRestoreChatHistory {
				log.Info("ENTER YOUR PIN TO RESTORE YOUR CHAT HISTORY")
				time.Sleep(1 * time.Second)
				continue
			}

			page = page.MustReload().MustWaitDOMStable()
			break
		}

		page = page.MustWaitLoad().MustWaitDOMStable()
		break
	}

	chats := page.MustElement(`div[aria-label="Chats"]`).MustElements(`div.x78zum5.xdt5ytf[data-virtualized]`)

	for _, chat := range chats {
		if strings.Contains(chat.MustText(), "Marketplace") {
			chat.MustClick()
			page = page.MustWaitLoad().MustWaitDOMStable()
			break
		}
	}

	marketplace := page.MustElement(`div[aria-label="Marketplace"]`)

	marketplace.MustHover()

	var hrefs []string

	for range 10 {
		log.Info("FETCHING MESSAGES")
		page.Mouse.MustScroll(0, 1000)
		page = page.MustWaitLoad().MustWaitDOMStable()
		chats = marketplace.MustElements(`div.x78zum5.xdt5ytf[data-virtualized]`)

		for _, chat := range chats {
			if chat.MustHas(`a[role="link"]`) {
				href := chat.MustElement(`a[role="link"]`).MustAttribute(`href`)
				if href != nil {
					hrefs = append(hrefs, *href)
				}
			}
		}
	}

	hrefs = removeDuplicates(hrefs)

	for _, href := range hrefs {
		href = fmt.Sprintf(("https://web.facebook.com%s"), href)
		page = page.MustNavigate(href).MustWaitLoad().MustWaitDOMStable()

		if !page.MustHas(`div.html-div.xexx8yu.x4uap5.x18d9i69.xkhd6sd.x1gslohp.x11i5rnm.x12nagc.x1mh8g0r.x1yc453h.x126k92a.xyk4ms5[dir="auto"]`) {

			messageInput := page.MustElement(`div[aria-label="Message"][contenteditable="true"]`)

			messageInput.MustInput(replyMessage)

			page.Keyboard.Press(input.Enter)
		}

		// Sleep for a random duration between 5 and 10 seconds
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		sleepDuration := 5 + r.Intn(6) // Random value between 5 and 10
		log.Info(fmt.Sprintf("Sleeping for %d seconds", sleepDuration))
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
}
