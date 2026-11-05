package oes

// Exercise holds the fields tools read. Documents keep every field on disk; this
// type is for checks and rendering, so it needn't round-trip extensions.
type Exercise struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Summary        string            `json:"summary,omitempty"`
	Status         string            `json:"status,omitempty"`
	ReplacedBy     string            `json:"replaced_by,omitempty"`
	PreviousIDs    []string          `json:"previous_ids,omitempty"`
	Aliases        []Alias           `json:"aliases,omitempty"`
	Classification Classification    `json:"classification"`
	Setup          map[string]any    `json:"setup,omitempty"`
	Anatomy        Anatomy           `json:"anatomy"`
	Equipment      Equipment         `json:"equipment"`
	Measurement    Measurement       `json:"measurement"`
	Instructions   Instructions      `json:"instructions"`
	Relations      []Relation        `json:"relations,omitempty"`
	ExternalIDs    map[string]string `json:"external_ids,omitempty"`
	Provenance     Provenance        `json:"provenance"`
}

type Alias struct {
	Text string `json:"text"`
	Lang string `json:"lang,omitempty"`
	Kind string `json:"kind,omitempty"`
}

type Classification struct {
	Modalities       []string `json:"modalities,omitempty"`
	MovementPatterns []string `json:"movement_patterns,omitempty"`
	Mechanics        string   `json:"mechanics,omitempty"`
	Force            string   `json:"force,omitempty"`
	KineticChain     string   `json:"kinetic_chain,omitempty"`
	Laterality       string   `json:"laterality,omitempty"`
	Planes           []string `json:"planes,omitempty"`
	Contraction      []string `json:"contraction,omitempty"`
	Level            string   `json:"level,omitempty"`
	Tags             []string `json:"tags,omitempty"`
}

type Anatomy struct {
	Muscles []MuscleUse `json:"muscles,omitempty"`
	Joints  []JointUse  `json:"joints,omitempty"`
}

type MuscleUse struct {
	Muscle string `json:"muscle"`
	Role   string `json:"role"`
}

type JointUse struct {
	Joint   string   `json:"joint"`
	Actions []string `json:"actions"`
}

type Equipment struct {
	Required [][]string `json:"required,omitempty"`
	Optional []string   `json:"optional,omitempty"`
}

type Measurement struct {
	Profiles        []Profile `json:"profiles"`
	Load            *Load     `json:"load,omitempty"`
	RepsCounted     string    `json:"reps_counted,omitempty"`
	DurationCounted string    `json:"duration_counted,omitempty"`
	Records         Records   `json:"records"`
}

type Profile struct {
	Required []string `json:"required"`
	Optional []string `json:"optional,omitempty"`
}

type Load struct {
	Kind       string `json:"kind"`
	Convention string `json:"convention,omitempty"`
	Implements int    `json:"implements,omitempty"`
}

type Records struct {
	E1RM   bool `json:"e1rm,omitempty"`
	Volume bool `json:"volume,omitempty"`
}

type Instructions struct {
	Setup        []string `json:"setup,omitempty"`
	Steps        []string `json:"steps,omitempty"`
	Breathing    string   `json:"breathing,omitempty"`
	Cues         []string `json:"cues,omitempty"`
	CommonFaults []Fault  `json:"common_faults,omitempty"`
	Safety       Safety   `json:"safety"`
}

type Fault struct {
	Fault      string `json:"fault"`
	Correction string `json:"correction"`
}

type Safety struct {
	Precautions []string `json:"precautions,omitempty"`
	Spotter     string   `json:"spotter,omitempty"`
	Safeties    bool     `json:"safeties,omitempty"`
}

type Relation struct {
	Type   string `json:"type"`
	Target string `json:"target"`
	Note   string `json:"note,omitempty"`
}

type Source struct {
	Title   string `json:"title"`
	URL     string `json:"url,omitempty"`
	License string `json:"license,omitempty"`
	Author  string `json:"author,omitempty"`
}

type Review struct {
	Status string   `json:"status"`
	Model  string   `json:"model,omitempty"`
	Models []string `json:"models,omitempty"`
	At     string   `json:"at,omitempty"`
	By     string   `json:"by,omitempty"`
}

type Provenance struct {
	License   string   `json:"license,omitempty"`
	Sources   []Source `json:"sources,omitempty"`
	Review    *Review  `json:"review,omitempty"`
	Generated []string `json:"generated,omitempty"`
}
