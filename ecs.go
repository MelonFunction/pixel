package main

import (
	"fmt"
	"log"
)

type EntityID uint32
type ComponentID uint32

type System interface {
	Update(dt float32)
	SetScene(scene *Scene)
}

type BasicSystem struct {
	System
	Scene *Scene
}

func (s *BasicSystem) SetScene(scene *Scene) {
	s.Scene = scene
}

type Entity struct {
	ID    EntityID
	Tag   Tag
	Scene *Scene
	Name  string // set this manually to help with debugging/printing
}

type Component struct {
	ID  ComponentID
	tag Tag

	entities   map[EntityID]interface{}
	destructor func(e *Entity, data interface{})
}

type QueryResult struct {
	Entity     *Entity
	Components map[*Component]interface{}
}

type Scene struct {
	currentEntityID    EntityID
	currentComponentID ComponentID

	entities    []*Entity
	entitiesMap map[EntityID]*Entity
	components  []*Component
	Systems     []System

	ComponentsMap map[string]*Component // allows for querying without a ref to the component
	Tags          map[string]Tag        // tags cache, allows searching for tag by string

}

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

	s.Tags[name] = t

	return t
}

func NewScene() *Scene {
	return &Scene{
		entities:      make([]*Entity, 0, 2000),
		entitiesMap:   make(map[EntityID]*Entity),
		ComponentsMap: make(map[string]*Component),
		components:    make([]*Component, 0, 64),
		Tags:          make(map[string]Tag),
	}
}

func (s *Scene) NewComponent(name string) *Component {
	c := &Component{
		ID: s.currentComponentID,
		tag: Tag{
			flags: 1 << s.currentComponentID,
		},
		entities: make(map[EntityID]interface{}),
	}
	s.currentComponentID++
	s.components = append(s.components, c)
	s.ComponentsMap[name] = c
	return c
}

// NewEntity creates and returns an Entity.
// If an EntityID is provided, it will use that instead of the automatic one.
// If an EntityID is provided and it  already exists, it will be removed and
// replaced, otherwise the ID will be skipped.
func (s *Scene) NewEntity(id *EntityID) *Entity {
	var e *Entity
	if id != nil {
		ID := *id
		if entity, ok := s.entitiesMap[ID]; ok {
			s.RemoveEntity(entity)
			fmt.Printf("replaced %d\n", ID)
		}
		e = &Entity{
			ID: ID,
		}
	} else {
		ok := true
		for ok {
			if _, ok = s.entitiesMap[s.currentEntityID]; ok {
				s.currentEntityID++
				fmt.Println("oops")
			}
		}

		e = &Entity{
			ID: s.currentEntityID,
		}
		s.currentEntityID++
	}
	s.entitiesMap[e.ID] = e
	s.entities = append(s.entities, e)
	e.Scene = s

	return e
}

func (s *Scene) AddSystem(sys System) *Scene {
	s.Systems = append(s.Systems, sys)
	sys.SetScene(s)
	return s
}

func (s *Scene) Update(dt float32) {
	for _, sys := range s.Systems {
		sys.Update(dt)
	}
}

func (s *Scene) Destroy() {
	cloned := s.entities[:]
	for i := 0; i < len(cloned); i++ {
		entity := cloned[len(cloned)-i-1]
		entity.Destroy()
	}
	log.Println("Destroyed, length of s.entities is", len(s.entities))
}

func (c *Component) SetDestructor(d func(e *Entity, data interface{})) {
	c.destructor = d
}

func (e *Entity) AddComponent(c *Component, data interface{}) *Entity {
	c.entities[e.ID] = data
	e.Tag.flags |= c.tag.flags
	return e
}

func (e *Entity) Destroy() {
	e.Scene.RemoveEntity(e)
}

func (e *Entity) RemoveComponent(c *Component) *Entity {
	if c.destructor != nil {
		if data, ok := c.entities[e.ID]; ok {
			c.destructor(e, data)
		}
	}
	delete(c.entities, e.ID)
	e.Tag.flags ^= c.tag.flags
	return e
}

// RemoveEntity removes an entity from the scene and also removes its component
// data (and calls the destructor if it was set)
func (s *Scene) RemoveEntity(e *Entity) {
	for i := 0; i < len(s.entities); i++ {
		entity := s.entities[i]
		if e.ID == entity.ID {
			for _, component := range s.components {
				if _, ok := component.entities[e.ID]; ok {
					entity.RemoveComponent(component)
				}
			}
			delete(s.entitiesMap, entity.ID)
			s.entities = append(s.entities[:i], s.entities[i+1:]...)
			break
		}
	}
}

// Query can accept a Tag or an EntityID. Multiple tags can be used which will
// include all entities which have that singlular tag. A composite tag made
// with s.BuildTag will exclude an entity if it's missing a component.
// TODO replace s.entities with s.taggedEntities or something similar
// TODO use multiple queries
func (s *Scene) QueryTag(tags ...Tag) []*QueryResult {
	ret := make([]*QueryResult, 0, 32)

	for _, entity := range s.entities {
		q := &QueryResult{
			Entity:     entity,
			Components: make(map[*Component]interface{}),
		}
		for _, tag := range tags {
			if entity.Tag.matches(tag) {
				for _, component := range s.components {
					if tag.matches(component.tag) { // t could be composite, so always bigger
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

	return ret
}

// QueryTagStrings converts strings into tags and then returns QueryTag's results
func (s *Scene) QueryTagStrings(tagStrings ...string) []*QueryResult {
	tags := make([]Tag, 0, 8)
	for _, tagString := range tagStrings {
		tag, ok := s.Tags[tagString]
		if !ok {
			panic("Tag doesn't exist: " + tagString)
		}
		tags = append(tags, tag)
	}
	return s.QueryTag(tags...)
}

// QueryID returns the result for a single EntityID
func (s *Scene) QueryID(id EntityID) (*QueryResult, error) {
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
