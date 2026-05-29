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