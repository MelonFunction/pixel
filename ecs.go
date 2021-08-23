package main

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
)

// System must have update and SetScene, but the latter is handled by
// BasicSystem if you bake it into a new system struct
type System interface {
	Update(dt float32)
	SetScene(scene *Scene)
}

// BasicSystem exists just for SetScene, just bake it into a new system struct
type BasicSystem struct {
	Scene *Scene
}

// SetScene sets the scene on a system
func (s *BasicSystem) SetScene(scene *Scene) {
	s.Scene = scene
}

// Entity can store multiple components
type Entity struct {
	ID    uint32
	Tag   Tag
	Scene *Scene
	Name  string // set this manually to help with debugging/printing
}

// Component is the base component
type Component struct {
	ID  uint32
	tag Tag

	*sync.RWMutex
	entities   map[uint32]interface{}
	destructor func(e *Entity, data interface{})
}

// QueryResult is the type which is returned when Scene.Query() is called
type QueryResult struct {
	Entity     *Entity
	Components map[*Component]interface{}
}

// Scene stores all of the entities, components and has functions to add and
// remove them
type Scene struct {
	currentEntityID    uint32
	currentComponentID uint32

	entities    []*Entity
	entitiesMap map[uint32]*Entity
	components  []*Component
	Systems     []System

	cache map[uint64][]*QueryResult

	*sync.RWMutex
	ComponentsMap map[string]*Component // allows for querying without a ref to the component
	Tags          map[string]Tag        // tags cache, allows searching for tag by string
}

// Tag is used to sort components
type Tag struct {
	name  string
	flags uint64 // Max 64
}

// matches checks if the comp Tag is present in t
func (t Tag) matches(comp Tag) bool {
	return t.flags&comp.flags == comp.flags
}

// BuildTag generates the tag and stores it in cache
func (s *Scene) BuildTag(name string, components ...interface{}) Tag {

	t := Tag{
		name: name,
	}
	for _, c := range components {
		switch typed := c.(type) {
		case *Component:
			t.flags |= typed.tag.flags
		case Tag:
			t.flags |= typed.flags
		default:
			panic("BuildTag only supports *Component or Tag types")
		}
	}

	s.Lock()
	defer s.Unlock()

	s.Tags[name] = t

	return t
}

// NewScene returns a new scene
func NewScene() *Scene {
	return &Scene{
		currentEntityID: 1,
		entities:        make([]*Entity, 0, 2000),
		entitiesMap:     make(map[uint32]*Entity),
		ComponentsMap:   make(map[string]*Component),
		components:      make([]*Component, 0, 64),
		Tags:            make(map[string]Tag),
		RWMutex:         &sync.RWMutex{},
		Systems:         make([]System, 0),
		cache:           make(map[uint64][]*QueryResult),
	}
}

// NewComponent returns a new component
func (s *Scene) NewComponent(name string) *Component {
	s.Lock()
	defer s.Unlock()

	c := &Component{
		ID: s.currentComponentID,
		tag: Tag{
			flags: 1 << s.currentComponentID,
		},
		entities: make(map[uint32]interface{}),
		RWMutex:  &sync.RWMutex{},
	}

	s.currentComponentID++
	s.components = append(s.components, c)
	s.ComponentsMap[name] = c
	return c
}

// NewEntity creates and returns an Entity.
// If the ID is non-zero, it will use that instead of the automatic one.
// If the ID is provided and it already exists, it will be removed and
// replaced
func (s *Scene) NewEntity(id uint32) *Entity {
	var e *Entity
	if id > 0 {
		// Use specified id
		s.RLock()
		oldEntity, ok := s.entitiesMap[id]
		s.RUnlock()
		if ok {
			s.RemoveEntity(oldEntity)
		}

		e = &Entity{
			ID: id,
		}
	} else {
		// loop until we get a free entity id
		var newID uint32
		for {
			s.RLock()
			_, ok := s.entitiesMap[s.currentEntityID]
			newID = s.currentEntityID
			s.RUnlock()
			if !ok {
				break
			}

			atomic.AddUint32(&s.currentEntityID, 1)
		}

		// log.Println("added new entity", newID, s.currentEntityID)
		e = &Entity{
			ID: newID,
		}
	}

	s.Lock()
	s.entitiesMap[e.ID] = e
	s.entities = append(s.entities, e)
	e.Scene = s
	s.Unlock()

	return e
}

// AddSystem adds the system to the scene, and the scene to the system
func (s *Scene) AddSystem(sys System) *Scene {
	s.Lock()
	defer s.Unlock()

	s.Systems = append(s.Systems, sys)
	sys.SetScene(s)
	return s
}

// Update updates all of the scenes
func (s *Scene) Update(dt float32) {
	for _, sys := range s.Systems {
		sys.Update(dt)
	}
}

// Destroy destroys the entities in the scene.
// It assumes that only a single thread is calling this and that no operations
// are currently happening on the Scene
func (s *Scene) Destroy() {
	cloned := s.entities[:]
	for i := 0; i < len(cloned); i++ {
		entity := cloned[len(cloned)-i-1]
		entity.Destroy()
	}
}

// SetDestructor sets the destructor on a component
func (c *Component) SetDestructor(d func(e *Entity, data interface{})) {
	c.destructor = d
}

// AddComponent adds a component to an entity, it returns itself for chaining
// A Component and a struct related to the data being stored should be created,
// ```
//
// var moveable = scene.NewComponent("moveable")
//
// type Moveable struct {
//     Bounds rl.Rectangle
// }
//
// e := scene.NewEntity(nil).
//		AddComponent(moveable, &Moveable{bounds, rl.Vector2{}})
// ````
func (e *Entity) AddComponent(c *Component, data interface{}) *Entity {
	c.Lock()
	defer c.Unlock()
	c.entities[e.ID] = data
	e.Tag.flags |= c.tag.flags

	// Remove cache entries which use this tag
	for tag := range e.Scene.cache {
		if tag&c.tag.flags == c.tag.flags {
			log.Println("deleted cache entry", tag)
			delete(e.Scene.cache, tag)
		}
	}

	return e
}

// Destroy removes the entity from the scene
func (e *Entity) Destroy() {
	e.Scene.RemoveEntity(e)
}

// RemoveComponent removes a component from the entity
func (e *Entity) RemoveComponent(c *Component) *Entity {
	if c.destructor != nil {
		c.RLock()
		if data, ok := c.entities[e.ID]; ok {
			c.destructor(e, data)
		}
		c.RUnlock()
	}
	c.Lock()
	delete(c.entities, e.ID)
	e.Tag.flags ^= c.tag.flags
	c.Unlock()
	return e
}

// RemoveEntity removes an entity from the scene and also removes its component
// data (and calls the destructor if it was set)
func (s *Scene) RemoveEntity(e *Entity) {
	s.Lock()
	defer s.Unlock()

	for i := 0; i < len(s.entities); i++ {
		entity := s.entities[i]
		if e.ID == entity.ID {
			for _, component := range s.components {
				component.RLock()
				_, ok := component.entities[e.ID]
				component.RUnlock()
				if ok {
					entity.RemoveComponent(component)
				}
			}
			delete(s.entitiesMap, entity.ID)
			s.entities = append(s.entities[:i], s.entities[i+1:]...)
			break
		}
	}
}

// QueryTag can accept a Tag or a uint32. Multiple tags can be used which will
// include all entities which have that singlular tag. A composite tag made
// with s.BuildTag will exclude an entity if it's missing a component.
// TODO replace s.entities with s.taggedEntities or something similar
// TODO use multiple queries
func (s *Scene) QueryTag(tags ...Tag) []*QueryResult {
	s.RLock()
	defer s.RUnlock()

	ret := make([]*QueryResult, 0, 32)

	// the tag used for this exact search
	var queryTag uint64
	for _, tag := range tags {
		queryTag |= tag.flags
	}

	// Return cached values if they exist
	if cached, ok := s.cache[queryTag]; ok {
		return cached
	}

	for _, entity := range s.entities {
		q := &QueryResult{
			Entity:     entity,
			Components: make(map[*Component]interface{}),
		}
		for _, tag := range tags {
			if entity.Tag.matches(tag) {
				for _, component := range s.components {
					if tag.matches(component.tag) {
						for e, v := range component.entities {
							if e == entity.ID {
								q.Components[component] = v
							}
						}
					}
				}
			}
		}
		if len(q.Components) > 0 {
			ret = append(ret, q)
		}
	}

	// Update cache
	log.Println("Not found, adding: ", queryTag, tags)
	s.cache[queryTag] = ret

	return ret
}

// QueryID returns the result for a single uint32
func (s *Scene) QueryID(id uint32) (*QueryResult, error) {
	s.RLock()
	defer s.RUnlock()

	entity, ok := s.entitiesMap[id]
	if ok {
		q := &QueryResult{
			Entity:     entity,
			Components: make(map[*Component]interface{}),
		}
		for _, component := range s.components {
			if entity.Tag.matches(component.tag) { // t could be composite, so always bigger
				for e, v := range component.entities {
					if e == entity.ID {
						q.Components[component] = v
					}
				}
			}
		}
		return q, nil
	}
	return nil, fmt.Errorf("Entity with ID %d not found", int(id))
}

// MoveEntityToEnd removes and reappends the entity
func (s *Scene) MoveEntityToEnd(entity *Entity) error {
	s.Lock()
	defer s.Unlock()

	found := false
	for i, e := range s.entities {
		if e == entity {
			found = true
			s.entities = append(s.entities[:i], s.entities[i+1:]...)
			break
		}
	}

	if found {
		s.entities = append(s.entities, entity)
	} else {
		return fmt.Errorf("Entity not found")
	}

	return nil
}
