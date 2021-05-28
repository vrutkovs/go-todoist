package todoist

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

type Section struct {
	Entity
	Name      string `json:"name"`
	ProjectID ID     `json:"project_id,omitempty"`
}

type NewSectionOpts struct {
	ParentID ID
}

func NewSection(name string, opts *NewSectionOpts) (*Section, error) {
	if len(name) == 0 {
		return nil, errors.New("new section requires a name")
	}
	section := Section{
		Name:      name,
		ProjectID: opts.ParentID,
	}
	section.ID = GenerateTempID()
	return &section, nil
}

func (s Section) String() string {
	return "#" + s.Name
}

type SectionClient struct {
	*Client
	cache *sectionCache
}

func (c *SectionClient) Add(section Section) (*Section, error) {
	c.cache.store(section)
	command := Command{
		Type:   "section_add",
		Args:   section,
		UUID:   GenerateUUID(),
		TempID: section.ID,
	}
	c.queue = append(c.queue, command)
	return &section, nil
}

func (c *SectionClient) Update(section Section) (*Section, error) {
	command := Command{
		Type: "section_update",
		Args: section,
		UUID: GenerateUUID(),
	}
	c.queue = append(c.queue, command)
	return &section, nil
}

func (c *SectionClient) Move(id, parentID ID) error {
	command := Command{
		Type: "section_move",
		UUID: GenerateUUID(),
		Args: map[string]ID{
			"id":        id,
			"parent_id": parentID,
		},
	}
	c.queue = append(c.queue, command)
	return nil

}

func (c *SectionClient) Delete(id ID) error {
	command := Command{
		Type: "section_delete",
		UUID: GenerateUUID(),
		Args: map[string]ID{
			"id": id,
		},
	}
	c.queue = append(c.queue, command)
	return nil
}

func (c *SectionClient) Archive(id ID) error {
	command := Command{
		Type: "section_archive",
		UUID: GenerateUUID(),
		Args: map[string]ID{
			"id": id,
		},
	}
	c.queue = append(c.queue, command)
	return nil
}

func (c *SectionClient) Unarchive(id ID) error {
	command := Command{
		Type: "section_unarchive",
		UUID: GenerateUUID(),
		Args: map[string]ID{
			"id": id,
		},
	}
	c.queue = append(c.queue, command)
	return nil
}

func (c *SectionClient) Reorder(projects []Section) error {
	command := Command{
		Type: "section_reorder",
		UUID: GenerateUUID(),
		Args: map[string][]Section{
			"projects": projects,
		},
	}
	c.queue = append(c.queue, command)
	return nil
}

type SectionGetResponse struct {
	Section Section
}

func (c *SectionClient) Get(ctx context.Context, id ID) (*SectionGetResponse, error) {
	values := url.Values{"section_id": {id.String()}}
	req, err := c.newRequest(ctx, http.MethodGet, "sections/get", values)
	if err != nil {
		return nil, err
	}
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	var out SectionGetResponse
	err = decodeBody(res, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *SectionClient) GetAll() []Section {
	return c.cache.getAll()
}

func (c *SectionClient) Resolve(id ID) *Section {
	return c.cache.resolve(id)
}

func (c SectionClient) FindByName(substr string) []Section {
	if r := []rune(substr); len(r) > 0 && string(r[0]) == "#" {
		substr = string(r[1:])
	}
	var res []Section
	for _, p := range c.GetAll() {
		if strings.Contains(p.Name, substr) {
			res = append(res, p)
		}
	}
	return res
}

func (c SectionClient) FindOneByName(substr string) *Section {
	projects := c.FindByName(substr)
	for _, project := range projects {
		if project.Name == substr {
			return &project
		}
	}
	if len(projects) > 0 {
		return &projects[0]
	}
	return nil
}

type sectionCache struct {
	cache *[]Section
}

func (c *sectionCache) getAll() []Section {
	return *c.cache
}

func (c *sectionCache) resolve(id ID) *Section {
	for _, section := range *c.cache {
		if section.ID == id {
			return &section
		}
	}
	return nil
}

func (c *sectionCache) store(section Section) {
	var res []Section
	isNew := true
	for _, s := range *c.cache {
		if s.Equal(section) {
			if !section.IsDeleted {
				res = append(res, section)
			}
			isNew = false
		} else {
			res = append(res, s)
		}
	}
	if isNew && !section.IsDeleted.Bool() {
		res = append(res, section)
	}
	c.cache = &res
}

func (c *sectionCache) remove(section Section) {
	var res []Section
	for _, s := range *c.cache {
		if !s.Equal(section) {
			res = append(res, s)
		}
	}
	c.cache = &res
}
