package osd

import "fmt"

const bufSize = 64

func proj(i int) int {
	return i % bufSize
}

type buffer struct {
	data       [bufSize]*event
	start, len int
}

func (b *buffer) first() *event {
	if b.len == 0 {
		return nil
	} else {
		return b.data[b.start]
	}
}

func (b *buffer) last() *event {
	if b.len == 0 {
		return nil
	} else {
		return b.data[proj(b.start+b.len-1)]
	}
}

func (b *buffer) push(e *event) {
	if b.len >= bufSize {
		fmt.Println("Push to full buffer, dropping event")
	} else {
		b.data[proj(b.start+b.len)] = e
		b.len++
	}

}

func (b *buffer) pop() {
	if b.len == 0 {
		fmt.Println("Pop from empty buffer")
	} else {
		b.data[b.start] = nil
		b.start = proj(b.start + 1)
		b.len--
	}
}

func (b *buffer) canMergeWithLast(e *event) bool {
	if b.last() == nil {
		return false
	} else if b.last().Sender != e.Sender {
		return false
	} else if b.last().Gauge != nil {
		return e.Gauge != nil
	} else {
		return b.last().Title == e.Title && len(b.last().Message) < 3
	}
}

func (b *buffer) mergeWithLast(e *event) {
	if b.last().Gauge != nil {
		b.data[proj(b.start+b.len-1)] = &event{
			Sender:   e.Sender,
			IconName: e.IconName,
			Gauge:    e.Gauge,
		}
	} else {
		b.data[proj(b.start+b.len-1)] = &event{
			Sender:   e.Sender,
			Title:    e.Title,
			Message:  append(e.Message, b.last().Message...),
			IconName: e.IconName,
		}
	}
}
