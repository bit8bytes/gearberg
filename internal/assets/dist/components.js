import { LitElement, html, nothing } from '/dist/lit-core.min.js';

// <combo-box> — accessible combobox web component
//
// Attributes:
//   input-id      — id placed on the visible <input> (for external <label for="…">)
//   name          — name for the hidden ID field (submitted when selecting an existing option)
//   name-new      — name for the hidden name field (submitted when typing a new value)
//   placeholder   — placeholder text for the visible input
//   selected-id   — pre-selected option value (server-side form re-population)
//   selected-name — display label matching selected-id (server-side form re-population)
//
// Children: <option value="id">Label</option> elements; consumed on first render.
//
// Keyboard:
//   ArrowDown / ArrowUp — move through visible options; ArrowDown also opens the list
//   Enter               — confirm the focused option
//   Escape              — close the list

let _uid = 0;

class ComboBox extends LitElement {
  static properties = {
    name:         { type: String },
    nameNew:      { type: String, attribute: 'name-new' },
    placeholder:  { type: String },
    inputid:      { type: String, attribute: 'input-id' },
    selectedId:   { type: String, attribute: 'selected-id' },
    selectedName: { type: String, attribute: 'selected-name' },
    _inputValue:     { state: true },
    _currentId:      { state: true },
    _currentNewName: { state: true },
    _isOpen:         { state: true },
    _activeIndex:    { state: true },
  };

  // Render into the element's own DOM so <label for> association and form
  // submission work without Shadow DOM or ElementInternals.
  createRenderRoot() { return this; }

  constructor() {
    super();
    this._uid = ++_uid;
    this._options = [];
    this.name = '';
    this.nameNew = '';
    this.placeholder = '';
    this.inputid = '';
    this.selectedId = '';
    this.selectedName = '';
    this._inputValue = '';
    this._currentId = '';
    this._currentNewName = '';
    this._isOpen = false;
    this._activeIndex = -1;
  }

  connectedCallback() {
    // Options live inside a <template> so they are inert until the component
    // upgrades — prevents raw text from flashing in browsers that render
    // <option> elements outside <select> (e.g. Chrome).
    const tmpl = this.querySelector('template');
    const opts = Array.from(tmpl ? tmpl.content.querySelectorAll('option') : []);
    this._options = opts.filter((o) => o.value !== '').map((o) => ({
      value: o.value,
      label: o.textContent.trim(),
    }));
    const sel = opts.find((o) => o.hasAttribute('selected'));
    if (sel) {
      this._currentId = sel.value;
      this._inputValue = sel.textContent.trim();
    } else if (this.selectedId && this.selectedName) {
      this._currentId = this.selectedId;
      this._inputValue = this.selectedName;
    } else if (!this.selectedId && this.selectedName) {
      // User previously typed a new name that has not been persisted yet.
      this._currentNewName = this.selectedName;
      this._inputValue = this.selectedName;
    }
    super.connectedCallback();
  }

  get _filtered() {
    return this._options;
  }

  get _showCreate() {
    const q = this._inputValue.trim();
    return q.length > 0 && !this._options.some((o) => o.label.toLowerCase() === q.toLowerCase());
  }

  // Ordered list of what is visible in the listbox.
  // The create option is always index 0 when present; existing options follow.
  get _visible() {
    return [
      ...(this._showCreate ? [{ value: '__create__', label: this._inputValue.trim() }] : []),
      ...this._filtered,
    ];
  }

  _optId(index) {
    return `combo-opt-${this._uid}-${index}`;
  }

  updated(changed) {
    // Scroll the keyboard-focused option into view after each render.
    if ((changed.has('_activeIndex') || changed.has('_isOpen')) &&
        this._isOpen && this._activeIndex >= 0) {
      this.querySelector(`#${this._optId(this._activeIndex)}`)
        ?.scrollIntoView({ block: 'nearest' });
    }
  }

  render() {
    const listId = `combo-list-${this._uid}`;
    const inputId = this.inputid || `combo-input-${this._uid}`;
    const showCreate = this._showCreate;
    const filtered = this._filtered;
    const activeDescendant = this._isOpen && this._activeIndex >= 0
      ? this._optId(this._activeIndex)
      : nothing;

    return html`
      <input type="hidden" name=${this.name} .value=${this._currentId}>
      <input type="hidden" name=${this.nameNew} .value=${this._currentNewName}>
      <div class="relative w-full">
        <input
          type="text"
          id=${inputId}
          class="input w-full"
          placeholder=${this.placeholder}
          autocomplete="off"
          role="combobox"
          aria-expanded=${this._isOpen ? 'true' : 'false'}
          aria-autocomplete="list"
          aria-haspopup="listbox"
          aria-controls=${listId}
          aria-activedescendant=${activeDescendant}
          .value=${this._inputValue}
          @input=${this._onInput}
          @focus=${this._onFocus}
          @click=${this._onClick}
          @keydown=${this._onKeydown}
          @blur=${this._onBlur}
        >
        <ul
          id=${listId}
          role="listbox"
          class="menu bg-base-100 border border-base-300 rounded-box absolute z-50 w-full mt-1 max-h-60 overflow-y-auto shadow ${this._isOpen ? '' : 'hidden'}"
        >
          ${showCreate ? html`
            <li
              id=${this._optId(0)}
              role="option"
              aria-selected="false"
            >
              <button type="button" class=${this._activeIndex === 0 ? 'menu-active' : ''} @mousedown=${this._onCreateMousedown}>
                + <span class="text-primary">${this._inputValue.trim()}</span>
              </button>
            </li>
          ` : nothing}
          ${filtered.map((opt, i) => {
            const idx = showCreate ? i + 1 : i;
            return html`
              <li
                id=${this._optId(idx)}
                role="option"
                aria-selected=${opt.value === this._currentId ? 'true' : 'false'}
              >
                <button
                  type="button"
                  class=${this._activeIndex === idx ? 'menu-active' : ''}
                  @mousedown=${() => this._selectExisting(opt.value, opt.label)}
                >
                  ${opt.label}
                </button>
              </li>
            `;
          })}
        </ul>
      </div>
    `;
  }

  _onInput(e) {
    this._inputValue = e.target.value;
    this._isOpen = this._options.length > 0 || this._inputValue.trim().length > 0;
    this._activeIndex = -1;
  }

  _onFocus() {
    this._isOpen = this._options.length > 0 || this._inputValue.trim().length > 0;
  }

  _onClick() {
    if (!this._isOpen) {
      this._isOpen = this._options.length > 0 || this._inputValue.trim().length > 0;
    }
  }

  _onKeydown(e) {
    const visible = this._visible;
    const isOpen = this._isOpen;

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (!isOpen && this._options.length > 0) {
        this._isOpen = true;
        this._activeIndex = 0;
      } else if (isOpen && visible.length > 0) {
        this._activeIndex = Math.min(this._activeIndex + 1, visible.length - 1);
      }
      return;
    }

    if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (isOpen) {
        this._activeIndex = Math.max(this._activeIndex - 1, 0);
      }
      return;
    }

    if (e.key === 'Enter' && isOpen && this._activeIndex >= 0) {
      e.preventDefault();
      const opt = visible[this._activeIndex];
      opt.value === '__create__'
        ? this._selectNew(opt.label)
        : this._selectExisting(opt.value, opt.label);
      return;
    }

    if (e.key === 'Escape') {
      this._isOpen = false;
      this._activeIndex = -1;
    }
  }

  _onBlur() {
    this._isOpen = false;
    this._activeIndex = -1;
    const typed = this._inputValue.trim();
    if (!typed) {
      this._currentId = '';
      this._currentNewName = '';
      return;
    }
    const exactMatch = this._options.find((o) => o.label.toLowerCase() === typed.toLowerCase());
    if (exactMatch) {
      this._selectExisting(exactMatch.value, exactMatch.label);
      return;
    }
    const matched = this._options.find((o) => o.value === this._currentId);
    if (!matched || matched.label !== typed) {
      this._selectNew(typed);
    }
  }

  _onCreateMousedown(e) {
    e.preventDefault(); // prevent blur firing before mousedown completes
    this._selectNew(this._inputValue.trim());
  }

  _selectExisting(id, label) {
    this._currentId = id;
    this._currentNewName = '';
    this._inputValue = label;
    this._isOpen = false;
    this._activeIndex = -1;
  }

  _selectNew(name) {
    this._currentId = '';
    this._currentNewName = name;
    this._inputValue = name;
    this._isOpen = false;
    this._activeIndex = -1;
  }
}

customElements.define('combo-box', ComboBox);
