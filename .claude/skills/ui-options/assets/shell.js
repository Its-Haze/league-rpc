// Renders each option's sample once per theme, tracks picks, and posts the
// result back to the preview server.
function renderSamples() {
  for (const option of document.querySelectorAll(".option")) {
    const template = option.querySelector("template.sample");
    if (!template) continue;
    const renders = document.createElement("div");
    renders.className = "renders";
    for (const [theme, label] of [
      ["theme-dark", "Dark"],
      ["theme-light", "Light"],
    ]) {
      const panel = document.createElement("div");
      panel.className = `panel ${theme}`;
      const tag = document.createElement("span");
      tag.className = "panel-label";
      tag.textContent = label;
      panel.append(tag, template.content.cloneNode(true));
      renders.append(panel);
    }
    template.replaceWith(renders);
  }
}

function buildFooters() {
  for (const option of document.querySelectorAll(".option")) {
    if (option.querySelector("footer")) continue;
    const footer = document.createElement("footer");
    const button = document.createElement("button");
    button.className = "pick";
    button.type = "button";
    button.textContent = `Pick ${option.dataset.option}`;
    footer.append(button);
    option.append(footer);
  }
}

const picks = {};

function wirePicks() {
  for (const option of document.querySelectorAll(".option")) {
    const decision = option.closest(".decision").dataset.decision;
    option.querySelector("button.pick").addEventListener("click", () => {
      const already = picks[decision] === option.dataset.option;
      for (const sibling of option.closest(".decision").querySelectorAll(".option")) {
        sibling.dataset.picked = "false";
      }
      if (already) {
        delete picks[decision];
      } else {
        picks[decision] = option.dataset.option;
        option.dataset.picked = "true";
      }
      setStatus("");
    });
  }
}

function wireViewToggle() {
  const buttons = document.querySelectorAll(".view-toggle button");
  for (const button of buttons) {
    button.addEventListener("click", () => {
      for (const other of buttons) {
        other.setAttribute("aria-pressed", String(other === button));
      }
      for (const renders of document.querySelectorAll(".renders")) {
        renders.dataset.view = button.dataset.view;
      }
    });
  }
}

function setStatus(text) {
  document.querySelector(".shell-foot .status").textContent = text;
}

function wireSend() {
  const notes = document.querySelector(".shell-foot textarea");
  const send = document.querySelector(".shell-foot button.send");
  send.addEventListener("click", async () => {
    const payload = { picks, notes: notes.value.trim(), at: new Date().toISOString() };
    try {
      const res = await fetch("/choice", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!res.ok) throw new Error(String(res.status));
      setStatus("Sent. Head back to Claude.");
    } catch {
      const blob = new Blob([JSON.stringify(payload, null, 2)], { type: "application/json" });
      const link = document.createElement("a");
      link.href = URL.createObjectURL(blob);
      link.download = "choice.json";
      link.click();
      setStatus("Server unreachable, downloaded choice.json instead.");
    }
  });
}

renderSamples();
buildFooters();
wirePicks();
wireViewToggle();
wireSend();
