package main

import (
	"fmt"
	"strings"

	"github.com/openexercise/open-exercise/internal/oes"
)

// answer is what the model returns for one source record.
type answer struct {
	IsExercise     bool               `json:"is_exercise"`
	Name           string             `json:"name"`
	Summary        string             `json:"summary"`
	Aliases        []oes.Alias        `json:"aliases"`
	Classification oes.Classification `json:"classification"`
	Position       string             `json:"position"`
	Incline        *float64           `json:"incline_degrees"`
	GripOrient     string             `json:"grip_orientation"`
	GripWidth      string             `json:"grip_width"`
	StanceWidth    string             `json:"stance_width"`
	StanceKind     string             `json:"stance_kind"`
	LoadPosition   string             `json:"load_position"`
	Range          string             `json:"range"`
	Muscles        []oes.MuscleUse    `json:"muscles"`
	Joints         []oes.JointUse     `json:"joints"`
	Equipment      []equipmentGroup   `json:"required_equipment"`
	Optional       []string           `json:"optional_equipment"`
	Profiles       []oes.Profile      `json:"measurement_profiles"`
	LoadKind       string             `json:"load_kind"`
	LoadConvention string             `json:"load_convention"`
	Implements     int                `json:"load_implements"`
	RepsCounted    string             `json:"reps_counted"`
	DurationCount  string             `json:"duration_counted"`
	E1RM           bool               `json:"records_e1rm"`
	Volume         bool               `json:"records_volume"`
	Instructions   oes.Instructions   `json:"instructions"`
}

type equipmentGroup struct {
	AnyOf []string `json:"any_of"`
}

// record builds an Open Exercise document. Empty model values are left out, not stored.
func (a answer) record(id string, relations []oes.Relation) map[string]any {
	setup := map[string]any{}
	put := func(m map[string]any, k, v string) {
		if v != "" {
			m[k] = v
		}
	}
	put(setup, "position", a.Position)
	put(setup, "load_position", a.LoadPosition)
	put(setup, "range", a.Range)
	if a.Incline != nil {
		setup["incline_degrees"] = *a.Incline
	}
	grip, stance := map[string]any{}, map[string]any{}
	put(grip, "orientation", a.GripOrient)
	put(grip, "width", a.GripWidth)
	put(stance, "width", a.StanceWidth)
	put(stance, "kind", a.StanceKind)
	if len(grip) > 0 {
		setup["grip"] = grip
	}
	if len(stance) > 0 {
		setup["stance"] = stance
	}
	required := make([][]string, 0, len(a.Equipment))
	for _, g := range a.Equipment {
		required = append(required, g.AnyOf)
	}
	equipment := map[string]any{"required": required}
	if len(a.Optional) > 0 {
		equipment["optional"] = a.Optional
	}
	load := map[string]any{"kind": a.LoadKind}
	put(load, "convention", a.LoadConvention)
	if a.Implements > 0 {
		load["implements"] = a.Implements
	}
	measurement := map[string]any{"profiles": a.Profiles, "load": load, "records": oes.Records{E1RM: a.E1RM, Volume: a.Volume}}
	put(measurement, "reps_counted", a.RepsCounted)
	put(measurement, "duration_counted", a.DurationCount)
	doc := map[string]any{
		"$schema": "https://openexercise.org/schema/v0/exercise.json",
		"id":      id, "name": a.Name, "summary": a.Summary, "status": "active",
		"classification": a.Classification,
		"anatomy":        oes.Anatomy{Muscles: a.Muscles, Joints: a.Joints},
		"equipment":      equipment, "measurement": measurement, "instructions": a.Instructions,
	}
	if len(setup) > 0 {
		doc["setup"] = setup
	}
	if len(relations) > 0 {
		doc["relations"] = relations
	}
	return doc
}

func describe(v oes.Vocabulary) string {
	var b strings.Builder
	for _, t := range v.Terms {
		fmt.Fprintf(&b, "%s", t.ID)
		if t.Parent != "" {
			fmt.Fprintf(&b, " (in %s)", t.Parent)
		}
		b.WriteString("; ")
	}
	return b.String()
}

func systemPrompt(v map[string]oes.Vocabulary) string {
	return `You are an exercise scientist and strength coach preparing one record for an open exercise catalogue. You get a community record which may be sparse, badly named or wrong. Produce an accurate record of the same exercise. Never turn it into a different exercise.

is_exercise: false only if the record is not a real exercise (a test entry, a whole routine, gibberish). Then leave everything else empty.

Name: the plain canonical English name in title case and singular, implement first when one defines it: "Barbell Back Squat", "Dumbbell Bench Press", "Cable Lateral Raise", "Pull-Up", "Plank". The same exercise must always get the same name, so use the most standard name, not a descriptive one. Keep facets that truly distinguish it (Incline, Single-Arm, Close-Grip).

Summary: one plain sentence.

Aliases: other names people really use, including any supplied. kind is common, abbreviation, regional, brand or misspelling. Never list the name itself.

Classification and setup: use the fixed values. Leave out a value that the exercise doesn't determine, or that you aren't sure of. List modalities and movement patterns most important first. level is technical demand.

Muscles: use the most specific term you can support, and don't list both a muscle and its parent. Only the muscles that matter, usually three to seven, each once; leave out incidental ones such as grip on a press or the hand muscles. prime-mover does most of the work (one to three; a muscle that merely assists is a synergist), synergist assists, stabilizer holds position, stretched only for stretches. The source's muscles may be wrong: correct them.

Joints: only those that move or are deliberately held under the working effort.

Equipment: required_equipment is a list of groups; all groups are needed, and any one item of a group is enough. Barbell squat: [{any_of:[barbell]},{any_of:[squat-rack,power-rack]}]. Nothing needed: an empty list. Don't include clothing or floor space.

Measurement: profiles are the ways to record a set, the usual one first. Each has required metrics and optional ones. Loaded lifts: required reps and load, optional rpe. Bodyweight reps: required reps. Holds and stretches: required duration. Cardio: required distance and duration, or duration; add what the machine or activity reports as optional.
load_kind says what the logged number means:
- external: the weight (barbell, dumbbell, cable, machine). Needs load_convention.
- bodyweight: no number logged, so no profile may contain load.
- bodyweight-plus-external: the number is what was added (weighted pull-up). Convention total.
- bodyweight-minus-assistance: the number is the assistance. Convention weight-stack, or total for a band.
- elastic: band tension; record band-tension, not load, unless the number is meant to be a weight.
- none: nothing is loaded (cardio, stretches), so no profile may contain load.
load_convention: total (everything moved, bar included), per-implement (one of two dumbbells or kettlebells held together, load_implements 2), per-side (one side of a plate-loaded machine or sled), weight-stack (the number on the pin; for cables and selectorized machines), not-applicable. load_implements: how many implements move at once; for a single dumbbell, 1 with convention total.
reps_counted: per-side if each side is done separately and reps are counted per side, else total. duration_counted the same for holds.
records_e1rm: true only for lifts with external load where an estimated one-rep max is meaningful (compound and isolation barbell, dumbbell, machine work with reps). False for bodyweight-based, ballistic, elastic and untimed work. records_volume: true when load times reps fairly measures work.

Instructions: write your own plain, specific text; do not copy the source. setup is how to get into position, steps one repetition or bout in order, breathing one sentence if it matters, cues two to four short coaching phrases, common_faults two to four with their corrections. safety.precautions only real ones; spotter required, recommended or not-needed; safeties true only when rack safeties apply. Do not give medical advice. Do not state numbers such as sets, reps, weights or percentages unless they define the movement.

Answer with one JSON object with these keys (leave out any that don't apply): is_exercise, name, summary, aliases [{text, kind}], classification {modalities [], movement_patterns [], mechanics: compound|isolation|not-applicable, force: push|pull|static|mixed|not-applicable, kinetic_chain: open|closed|mixed|not-applicable, laterality: bilateral|unilateral|alternating|contralateral|not-applicable, planes: [sagittal|frontal|transverse], contraction: [dynamic|isometric|eccentric-emphasis|ballistic], level: beginner|intermediate|advanced}, position: standing|seated|supine|prone|side-lying|kneeling|half-kneeling|quadruped|hanging|support|split-stance|inverted|bent-over|moving, incline_degrees, grip_orientation: pronated|supinated|neutral|mixed|hook|false|rotating, grip_width: narrow|shoulder|wide|snatch, stance_width: narrow|hip|shoulder|wide|sumo, stance_kind: parallel|staggered|split|single-leg|tandem, load_position: high-bar|low-bar|front-rack|overhead|goblet|suitcase|racked|at-sides|hip|chest|back|feet|zercher, range: full|partial|pause|deficit|pin|box, muscles [{muscle, role}], joints [{joint, actions []}], required_equipment [{any_of []}], optional_equipment [], measurement_profiles [{required [], optional []}], load_kind, load_convention, load_implements, reps_counted, duration_counted, records_e1rm, records_volume, instructions {setup [], steps [], breathing, cues [], common_faults [{fault, correction}], safety {precautions [], spotter, safeties}}.

Muscles: ` + describe(v["muscles"]) + "\nEquipment: " + describe(v["equipment"]) + "\nMovement patterns: " + describe(v["movement-patterns"]) + "\nModalities: " + describe(v["modalities"]) + "\nJoints: " + describe(v["joints"]) + "\nJoint actions: " + describe(v["joint-actions"]) + "\nMetrics: " + describe(v["metrics"]) + "\nUse only these exact terms."
}
