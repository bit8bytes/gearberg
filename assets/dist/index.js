// History back
// Trigger: <button data-history-back>
document.addEventListener("click", (e) => {
  if (e.target.closest("[data-history-back]")) {
    history.back();
  }
});