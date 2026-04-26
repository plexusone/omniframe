package omniframe

import (
	"errors"
	"fmt"
)

var (
	ErrFrameNotFound = errors.New("frame not found")
	ErrFrameExists   = errors.New("frame already exists")
	ErrEmptyFrameSet = errors.New("frame set is empty")
)

// FrameSet represents a collection of frames, useful for multi-sheet workbooks.
type FrameSet struct {
	name   string
	frames map[string]*Frame
	order  []string // Frame names in order
}

// NewFrameSet creates a new empty frame set.
func NewFrameSet(name string) *FrameSet {
	return &FrameSet{
		name:   name,
		frames: make(map[string]*Frame),
		order:  []string{},
	}
}

// Name returns the frame set name.
func (fs *FrameSet) Name() string {
	return fs.name
}

// SetName sets the frame set name.
func (fs *FrameSet) SetName(name string) {
	fs.name = name
}

// Len returns the number of frames.
func (fs *FrameSet) Len() int {
	return len(fs.order)
}

// Names returns the frame names in order.
func (fs *FrameSet) Names() []string {
	result := make([]string, len(fs.order))
	copy(result, fs.order)
	return result
}

// AddFrame adds a frame to the set using the frame's name.
// Returns an error if a frame with the same name already exists.
func (fs *FrameSet) AddFrame(f *Frame) error {
	if f == nil {
		return errors.New("frame cannot be nil")
	}

	name := f.Name()
	if name == "" {
		name = fmt.Sprintf("Sheet%d", len(fs.frames)+1)
		f.SetName(name)
	}

	if _, ok := fs.frames[name]; ok {
		return fmt.Errorf("%w: %s", ErrFrameExists, name)
	}

	fs.frames[name] = f
	fs.order = append(fs.order, name)
	return nil
}

// AddFrameAs adds a frame with a custom name.
func (fs *FrameSet) AddFrameAs(f *Frame, name string) error {
	if f == nil {
		return errors.New("frame cannot be nil")
	}

	if _, ok := fs.frames[name]; ok {
		return fmt.Errorf("%w: %s", ErrFrameExists, name)
	}

	// Clone the frame with the new name
	clone := f.Clone()
	clone.SetName(name)

	fs.frames[name] = clone
	fs.order = append(fs.order, name)
	return nil
}

// Frame returns the frame with the given name.
func (fs *FrameSet) Frame(name string) (*Frame, error) {
	f, ok := fs.frames[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrFrameNotFound, name)
	}
	return f, nil
}

// FrameAt returns the frame at the given index.
func (fs *FrameSet) FrameAt(idx int) (*Frame, error) {
	if idx < 0 || idx >= len(fs.order) {
		return nil, fmt.Errorf("index out of range: %d", idx)
	}
	return fs.frames[fs.order[idx]], nil
}

// Frames returns all frames in order.
func (fs *FrameSet) Frames() []*Frame {
	result := make([]*Frame, len(fs.order))
	for i, name := range fs.order {
		result[i] = fs.frames[name]
	}
	return result
}

// Remove removes a frame from the set.
func (fs *FrameSet) Remove(name string) error {
	if _, ok := fs.frames[name]; !ok {
		return fmt.Errorf("%w: %s", ErrFrameNotFound, name)
	}

	delete(fs.frames, name)

	// Remove from order
	newOrder := make([]string, 0, len(fs.order)-1)
	for _, n := range fs.order {
		if n != name {
			newOrder = append(newOrder, n)
		}
	}
	fs.order = newOrder

	return nil
}

// Reorder sets the order of frames.
func (fs *FrameSet) Reorder(names []string) error {
	// Validate all names exist
	for _, name := range names {
		if _, ok := fs.frames[name]; !ok {
			return fmt.Errorf("%w: %s", ErrFrameNotFound, name)
		}
	}

	// Validate no duplicates and all frames included
	seen := make(map[string]bool)
	for _, name := range names {
		if seen[name] {
			return fmt.Errorf("duplicate frame name: %s", name)
		}
		seen[name] = true
	}

	if len(names) != len(fs.frames) {
		return fmt.Errorf("must include all frames: expected %d, got %d",
			len(fs.frames), len(names))
	}

	fs.order = names
	return nil
}
