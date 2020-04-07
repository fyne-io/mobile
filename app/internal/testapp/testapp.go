// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux

// Small test app used by app/app_test.go.
package main

import (
	"log"
	"net"

	"github.com/fyne-io/mobile/app"
	"github.com/fyne-io/mobile/app/internal/apptest"
	"github.com/fyne-io/mobile/event/lifecycle"
	"github.com/fyne-io/mobile/event/paint"
	"github.com/fyne-io/mobile/event/size"
	"github.com/fyne-io/mobile/event/touch"
	"github.com/fyne-io/mobile/gl"
)

func main() {
	app.Main(func(a app.App) {
		var (
			glctx   gl.Context
			visible bool
		)

		addr := "127.0.0.1:" + apptest.Port
		log.Printf("addr: %s", addr)

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		log.Printf("dialled")
		comm := &apptest.Comm{
			Conn:   conn,
			Fatalf: log.Panicf,
			Printf: log.Printf,
		}

		comm.Send("hello_from_testapp")
		comm.Recv("hello_from_host")

		color := "red"
		sendPainting := false
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					comm.Send("lifecycle_visible")
					sendPainting = true
					visible = true
					glctx, _ = e.DrawContext.(gl.Context)
				case lifecycle.CrossOff:
					comm.Send("lifecycle_not_visible")
					visible = false
				}
			case size.Event:
				comm.Send("size", e.PixelsPerPt, e.Orientation)
			case paint.Event:
				if visible {
					if color == "red" {
						glctx.ClearColor(1, 0, 0, 1)
					} else {
						glctx.ClearColor(0, 1, 0, 1)
					}
					glctx.Clear(gl.COLOR_BUFFER_BIT)
					a.Publish()
				}
				if sendPainting {
					comm.Send("paint", color)
					sendPainting = false
				}
			case touch.Event:
				comm.Send("touch", e.Type, e.X, e.Y)
				if e.Type == touch.TypeEnd {
					if color == "red" {
						color = "green"
					} else {
						color = "red"
					}
					sendPainting = true
					// Send a paint event so the screen gets redrawn.
					a.Send(paint.Event{})
				}
			}
		}
	})
}
