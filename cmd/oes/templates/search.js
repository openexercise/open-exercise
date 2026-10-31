(async () => {
  const res = await fetch("/catalogue/index.json");
  const index = await res.json();
  const byID = new Map(index.map((e) => [e.id, e]));
  const q = document.getElementById("q");
  const pattern = document.getElementById("pattern");
  const muscle = document.getElementById("muscle");
  const equipment = document.getElementById("equipment");
  const count = document.getElementById("count");
  const rows = [...document.querySelectorAll("#results li")];
  const has = (list, v) => !v || (list || []).includes(v);
  const apply = () => {
    const words = q.value.toLowerCase().split(/\s+/).filter(Boolean);
    let shown = 0;
    for (const row of rows) {
      const e = byID.get(row.dataset.id);
      const text = [e.name, e.id, ...(e.aliases || [])].join(" ").toLowerCase();
      const match = words.every((w) => text.includes(w)) && has(e.patterns, pattern.value) && has(e.muscles, muscle.value) && has(e.equipment, equipment.value);
      row.hidden = !match;
      if (match) shown++;
    }
    count.textContent = shown === 1 ? "1 exercise" : shown + " exercises";
  };
  document.getElementById("filters").addEventListener("input", apply);
  document.getElementById("filters").addEventListener("submit", (ev) => ev.preventDefault());
  apply();
})();
