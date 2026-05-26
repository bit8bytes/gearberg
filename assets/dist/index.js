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