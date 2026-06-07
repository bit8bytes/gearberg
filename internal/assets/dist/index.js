// Loading state on form submit
// Usage: <button data-loading-trigger> with optional <span data-loading-spinner> inside
document.addEventListener("submit", (e) => {
  const trigger = e.target.querySelector("[data-loading-trigger]");
  if (!trigger) return;
  const spinner = trigger.querySelector("[data-loading-spinner]");
  if (spinner) spinner.classList.remove("hidden");
  trigger.disabled = true;
});

// Confirm before destructive form submissions
// Trigger: <button data-confirm="Are you sure?">
document.addEventListener("click", (e) => {
  const btn = e.target.closest("[data-confirm]");
  if (!btn) return;
  if (!confirm(btn.dataset.confirm)) {
    e.preventDefault();
  }
});

// History back
// Trigger: <button data-history-back>
document.addEventListener("click", (e) => {
  if (e.target.closest("[data-history-back]")) {
    history.back();
  }
});

// Auto-submit filter form on checkbox change
// Trigger: <form data-auto-submit> containing checkboxes
document.addEventListener("change", (e) => {
  const form = e.target.closest("form[data-auto-submit]");
  if (form) {
    form.submit();
  }
});

// Open a <dialog> modal
// Trigger: <button data-modal-target="dialog-id">
document.addEventListener("click", (e) => {
  const btn = e.target.closest("[data-modal-target]");
  if (!btn) return;
  const modal = document.getElementById(btn.dataset.modalTarget);
  if (modal) modal.showModal();
});

// In-place image preview
// Trigger: <input data-image-input> inside a container with <label data-image-preview>
document.addEventListener("change", (e) => {
  const input = e.target.closest("[data-image-input]");
  if (!input || !input.files[0]) return;
  const preview = input.closest("form").querySelector("[data-image-preview]");
  if (!preview) return;
  const img = preview.querySelector("[data-image-preview-img]");
  const placeholder = preview.querySelector("[data-image-placeholder]");
  const reader = new FileReader();
  reader.onload = (ev) => {
    img.src = ev.target.result;
    img.classList.remove("hidden");
    if (placeholder) placeholder.classList.add("hidden");
  };
  reader.readAsDataURL(input.files[0]);
});

// Type card selector + adaptive count label
// Trigger: radio inputs inside [data-type-option] labels within [data-type-selector]
(function () {
  function syncTypeCards(selector) {
    selector.querySelectorAll("[data-type-option]").forEach((card) => {
      const radio = card.querySelector("input[type='radio']");
      const checked = radio && radio.checked;
      card.classList.toggle("border-primary", checked);
      card.classList.toggle("bg-primary/10", checked);
      card.classList.toggle("border-base-300", !checked);
      card.classList.toggle("bg-base-200", !checked);
    });
  }

  function syncCountLabel(typeValue) {
    const label = document.querySelector("[data-count-label]");
    const hint = document.querySelector("[data-count-hint]");
    if (label) label.textContent = typeValue === "serialized" ? "Units to Create" : "Quantity";
    if (hint) hint.classList.toggle("hidden", typeValue !== "serialized");
  }

  document.addEventListener("change", (e) => {
    const card = e.target.closest("[data-type-option]");
    if (!card) return;
    const selector = card.closest("[data-type-selector]");
    if (selector) syncTypeCards(selector);
    syncCountLabel(e.target.value);
  });

  document.addEventListener("DOMContentLoaded", () => {
    const selector = document.querySelector("[data-type-selector]");
    if (!selector) return;
    syncTypeCards(selector);
    const checked = selector.querySelector("input[type='radio']:checked");
    if (checked) syncCountLabel(checked.value);
  });
})();


// Serialized inventory unit rows
// Trigger: [data-add-unit] adds a row; [data-remove-unit] removes its row.
// Unit numbers in .unit-number cells are kept in sync after every change.
(function () {
  function renumberRows(tbody) {
    tbody.querySelectorAll("tr").forEach((row, i) => {
      const cell = row.querySelector(".unit-number");
      if (cell) cell.textContent = i + 1;
    });
  }

  function addRow(tbody) {
    const template = document.getElementById("unit-row-template");
    if (!template) return;
    const clone = template.content.cloneNode(true);
    clone.querySelectorAll("input").forEach((input) => (input.value = ""));
    tbody.appendChild(clone);
    renumberRows(tbody);
  }

  document.addEventListener("click", (e) => {
    // Add unit row
    if (e.target.closest("[data-add-unit]")) {
      const tbody = document.getElementById("unit-rows");
      if (tbody) addRow(tbody);
      return;
    }

    // Remove unit row
    const removeBtn = e.target.closest("[data-remove-unit]");
    if (removeBtn) {
      const tbody = document.getElementById("unit-rows");
      const row = removeBtn.closest("tr");
      // Keep at least one row
      if (tbody.querySelectorAll("tr").length > 1) {
        row.remove();
        renumberRows(tbody);
      }
    }
  });

  // Auto-add the first row when the serialized form is present
  document.addEventListener("DOMContentLoaded", () => {
    const tbody = document.getElementById("unit-rows");
    if (!tbody) return;
    if (tbody.querySelectorAll("tr").length === 0) addRow(tbody);
  });
})();