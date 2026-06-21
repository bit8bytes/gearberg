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
// Trigger: <input data-image-input> inside a form with [data-image-preview]
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

// Photo expand button visibility
// Show the expand button whenever [data-image-preview-img] has a non-empty src.
function syncExpandButton() {
  const img = document.querySelector("[data-image-preview-img]");
  const btn = document.querySelector("[data-photo-expand]");
  if (!btn) return;
  const hasSrc = img && img.src && !img.src.endsWith(window.location.href);
  btn.classList.toggle("hidden", !hasSrc);
}
document.addEventListener("DOMContentLoaded", syncExpandButton);
document.addEventListener("change", (e) => {
  if (e.target.closest("[data-image-input]") || e.target.closest("[data-camera-input]")) {
    // Preview update happens asynchronously via FileReader; check after it settles.
    setTimeout(syncExpandButton, 50);
  }
});

// Photo viewer – copy current preview src into the viewer dialog before it opens.
document.addEventListener("click", (e) => {
  if (!e.target.closest('[data-modal-target="photo-viewer-dialog"]')) return;
  const previewImg = document.querySelector("[data-image-preview-img]");
  const viewerImg = document.querySelector("[data-photo-viewer-img]");
  if (previewImg && viewerImg) viewerImg.src = previewImg.src;
});

// Photo picker dialog
// Trigger: <button data-photo-option="camera|upload|url"> inside #photo-picker-dialog
(function () {
  function getForm(dialog) {
    return document.querySelector("form[enctype='multipart/form-data']");
  }

  document.addEventListener("click", (e) => {
    const opt = e.target.closest("[data-photo-option]");
    if (!opt) return;

    const dialog = opt.closest("dialog");
    const form = getForm(dialog);
    const imageInput = form && form.querySelector("[data-image-input]");
    const cameraInput = form && form.querySelector("[data-camera-input]");
    const action = opt.dataset.photoOption;

    if (action === "camera" && cameraInput) {
      dialog.close();
      cameraInput.click();
    } else if (action === "upload" && imageInput) {
      dialog.close();
      imageInput.click();
    } else if (action === "url") {
      const urlSection = dialog.querySelector("[data-url-input-section]");
      if (urlSection) {
        urlSection.classList.remove("hidden");
        const urlInput = urlSection.querySelector("[data-photo-url-input]");
        if (urlInput) urlInput.focus();
      }
    }
  });

  // Reset URL section when dialog closes
  document.addEventListener("close", (e) => {
    const dialog = e.target.closest && e.target.closest("dialog") || (e.target.tagName === "DIALOG" ? e.target : null);
    if (!dialog) return;
    const urlSection = dialog.querySelector("[data-url-input-section]");
    if (urlSection) {
      urlSection.classList.add("hidden");
      const urlInput = urlSection.querySelector("[data-photo-url-input]");
      if (urlInput) urlInput.value = "";
      const errorEl = urlSection.querySelector("[data-photo-url-error]");
      if (errorEl) errorEl.classList.add("hidden");
    }
  }, true);

  // Camera input → transfer file to main image input and trigger preview update
  document.addEventListener("change", (e) => {
    const cameraInput = e.target.closest("[data-camera-input]");
    if (!cameraInput || !cameraInput.files[0]) return;
    const form = cameraInput.closest("form");
    const imageInput = form && form.querySelector("[data-image-input]");
    if (!imageInput) return;
    const dt = new DataTransfer();
    dt.items.add(cameraInput.files[0]);
    imageInput.files = dt.files;
    imageInput.dispatchEvent(new Event("change", { bubbles: true }));
    cameraInput.value = "";
  });

  // URL download → fetch via server proxy, set file on image input
  document.addEventListener("click", async (e) => {
    const btn = e.target.closest("[data-photo-url-submit]");
    if (!btn) return;

    const urlSection = btn.closest("[data-url-input-section]");
    const dialog = btn.closest("dialog");
    const urlInput = urlSection && urlSection.querySelector("[data-photo-url-input]");
    const errorEl = urlSection && urlSection.querySelector("[data-photo-url-error]");
    const form = document.querySelector("form[enctype='multipart/form-data']");
    const imageInput = form && form.querySelector("[data-image-input]");

    const imageURL = urlInput ? urlInput.value.trim() : "";
    if (!imageURL || !imageInput) return;

    if (errorEl) errorEl.classList.add("hidden");
    btn.disabled = true;
    const originalText = btn.textContent;
    btn.textContent = "Downloading…";

    try {
      const res = await fetch("/image-proxy?url=" + encodeURIComponent(imageURL));
      if (!res.ok) {
        throw new Error((await res.text()).trim() || "Download failed.");
      }
      const blob = await res.blob();
      const filename = imageURL.split("/").pop().split("?")[0] || "image.jpg";
      const file = new File([blob], filename, { type: blob.type });
      const dt = new DataTransfer();
      dt.items.add(file);
      imageInput.files = dt.files;
      imageInput.dispatchEvent(new Event("change", { bubbles: true }));
      if (dialog) dialog.close();
    } catch (err) {
      if (errorEl) {
        errorEl.textContent = err.message;
        errorEl.classList.remove("hidden");
      }
    } finally {
      btn.disabled = false;
      btn.textContent = originalText;
    }
  });
})();

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

  function syncHasContent(typeValue) {
    const row = document.querySelector("[data-has-content-row]");
    const toggle = document.querySelector("[data-has-content-toggle]");
    if (!row) return;
    const visible = typeValue === "serialized";
    row.style.display = visible ? "" : "none";
    if (!visible && toggle) toggle.checked = false;
  }

  document.addEventListener("change", (e) => {
    const card = e.target.closest("[data-type-option]");
    if (!card) return;
    const selector = card.closest("[data-type-selector]");
    if (selector) syncTypeCards(selector);
    syncCountLabel(e.target.value);
    syncHasContent(e.target.value);
  });

  document.addEventListener("DOMContentLoaded", () => {
    const selector = document.querySelector("[data-type-selector]");
    if (!selector) return;
    syncTypeCards(selector);
    const checked = selector.querySelector("input[type='radio']:checked");
    if (checked) {
      syncCountLabel(checked.value);
      syncHasContent(checked.value);
    }
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