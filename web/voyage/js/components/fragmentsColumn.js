import {
  html,
  render,
} from "https://cdn.jsdelivr.net/gh/lit/dist@3/all/lit-all.min.js";
import { showConfirmation } from "../utils.js";
import {
  addFragment,
  deleteFragment,
  toggleCheckedFragment,
  unselectAllFragments,
  saveSelection,
  deleteSavedSelection,
} from "../slices/promptFragmentsSlice.js";
import {
  setCurrentPrompt,
} from "../slices/promptHistorySlice.js";


class FragmentsColumn {
  constructor(store, updateUI) {
    this.store = store;
    this.updateUI = updateUI;
    this.element = document.getElementById("fragments-column");
    this.init();
  }

  init() {
    this.element.addEventListener("click", (e) => {
      if (e.target.id === "add-fragment-btn") this.addFragment();
      if (e.target.id === "randomize-btn") this.randomizeAndAddFragments();
      if (e.target.id === "unselect-all-btn") this.unselectAllFragments();
      if (e.target.id === "save-selection-btn") this.openSaveSelectionModal();
    });
  }

  render() {
    const state = this.store.getState();
    const fragments = state.promptFragments.prompt_fragments || [];
    const checkedFragments = state.promptFragments.checked_fragments || [];
    const currentPrompt = state.promptHistory.current_prompt || "";
    const savedSelections = state.promptFragments.saved_selections || [];

    const template = html`
      <h2>Prompt Fragments</h2>
      <div class="checkbox-group" id="fragments-list">
        ${fragments.map((fragment, index) =>
          this.renderFragment(fragment, index, checkedFragments, currentPrompt)
        )}
      </div>
      <button id="add-fragment-btn">Add New Fragment</button>
      <div class="button-group">
        <button id="randomize-btn" class="randomize-btn">Randomize</button>
        <button id="unselect-all-btn">Unselect All</button>
      </div>
      <button id="save-selection-btn">Save Fragment Selection</button>
      <h3>Saved Selections</h3>
      <div id="saved-selections-list">
        ${savedSelections.map((savedSelection, index) =>
          this.renderSavedSelection(savedSelection, index)
        )}
      </div>
    `;

    render(template, this.element);
  }

  renderFragment(fragment, index, checkedFragments, currentPrompt) {
    return html`
      <div class="list-item">
        <input
          type="checkbox"
          id="fragment-${index}"
          ?checked=${checkedFragments.includes(index)}
          @change=${() => this.store.dispatch(toggleCheckedFragment(index))}
        />
        <label
          for="fragment-${index}"
          class=${this.isFragmentInPrompt(fragment, currentPrompt)
            ? "active-fragment"
            : ""}
          @click=${(e) => {
            e.preventDefault();
            this.toggleFragment(fragment);
          }}
        >
          ${fragment}
        </label>
        <button @click=${() => this.store.dispatch(deleteFragment(index))}>
          Delete
        </button>
      </div>
    `;
  }

  renderSavedSelection(savedSelection, index) {
    return html`
      <div class="list-item">
        <span
          @click=${() => this.restoreSavedSelection(savedSelection.selection)}
          >${savedSelection.name}</span
        >
        <button
          @click=${() => this.store.dispatch(deleteSavedSelection(index))}
        >
          Delete
        </button>
      </div>
    `;
  }

  isFragmentInPrompt(fragment, prompt) {
    return prompt.includes(fragment);
  }

  toggleFragment(fragment) {
    const state = this.store.getState();
    let currentPrompt = state.promptHistory.current_prompt || "";

    if (this.isFragmentInPrompt(fragment, currentPrompt)) {
      currentPrompt = this.removeFragmentFromPrompt(fragment, currentPrompt);
    } else {
      currentPrompt = this.addFragmentToPrompt(fragment, currentPrompt);
    }

    this.store.dispatch(setCurrentPrompt(currentPrompt.trim()));
    this.updateUI();
    showConfirmation(`Fragment "${fragment}" toggled`);
  }

  addFragmentToPrompt(fragment, prompt) {
    return prompt ? `${prompt}, ${fragment}` : fragment;
  }

  removeFragmentFromPrompt(fragment, prompt) {
    const regex = new RegExp(
      `(,\\s*)?${this.escapeRegExp(fragment)}(,\\s*)?`,
      "g"
    );
    let newPrompt = prompt.replace(regex, ",");
    newPrompt = newPrompt.replace(/^,\s*/, "").replace(/,\s*$/, "");
    return newPrompt;
  }

  escapeRegExp(string) {
    return string.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  }

  addFragment() {
    const fragment = prompt("Enter new prompt fragment:");
    if (fragment) {
      this.store.dispatch(addFragment(fragment.trim()));
      this.updateUI();
      showConfirmation("Fragment added successfully!");
    }
  }

  async randomizeAndAddFragments() {
    const state = this.store.getState();
    const fragments = state.promptFragments.prompt_fragments;
    const checkedFragments = state.promptFragments.checked_fragments;
    const selectedFragments = checkedFragments.map((index) => fragments[index]);

    log.debug("Selected fragments:", selectedFragments);

    if (selectedFragments.length === 0) {
      log.debug("No fragments selected.");
      return;
    }

    let currentPrompt = state.promptHistory.current_prompt || "";
    const currentFragments = currentPrompt.split(",").map((f) => f.trim());

    log.debug("Current prompt before removal:", currentPrompt);

    // Remove all selected fragments from the current prompt
    selectedFragments.forEach((fragment) => {
      currentPrompt = this.removeFragmentFromPrompt(fragment, currentPrompt);
    });

    log.debug("Current prompt after removal:", currentPrompt);

    const availableFragments = selectedFragments.filter(
      (f) => !currentFragments.includes(f)
    );

    if (availableFragments.length === 0) {
      showConfirmation("All selected fragments are already in the prompt!");
      log.debug("All selected fragments are already in the prompt.");
      return;
    }

    const numberToSelect = Math.min(
      Math.floor(Math.random() * availableFragments.length) + 1,
      availableFragments.length
    );
    const shuffled = availableFragments.sort(() => 0.5 - Math.random());
    const randomizedFragments = shuffled.slice(0, numberToSelect);
    const fragmentsText = randomizedFragments.join(", ");

    log.debug("Randomized fragments to add:", randomizedFragments);

    const newPrompt = currentPrompt
      ? `${currentPrompt}, ${fragmentsText}`
      : fragmentsText;
    this.store.dispatch(setCurrentPrompt(newPrompt.trim()));
    this.updateUI();
    showConfirmation("Random fragments added to prompt!");

    log.debug("New prompt:", newPrompt);
  }

  unselectAllFragments() {
    this.store.dispatch(unselectAllFragments());
    this.updateUI();
    showConfirmation("All fragments unselected!");
  }

  openSaveSelectionModal() {
    const modal = document.getElementById("save-selection-modal");
    modal.style.display = "flex";
    document.getElementById("selection-name").value = "";
    document.getElementById("selection-name").focus();

    if (!this.saveSelectionListener) {
      this.saveSelectionListener = () => this.saveFragmentSelection();
      document
        .getElementById("confirm-save-selection-btn")
        .addEventListener("click", this.saveSelectionListener);
      document
        .getElementById("cancel-save-selection-btn")
        .addEventListener("click", () => this.closeSaveSelectionModal());
    }
  }

  saveFragmentSelection() {
    const name = document.getElementById("selection-name").value.trim();
    if (name) {
      const state = this.store.getState();
      const checkedFragments = state.promptFragments.checked_fragments;
      const selection = { name, selection: checkedFragments };
      this.store.dispatch(saveSelection(selection));
      this.updateUI();
      this.closeSaveSelectionModal();
      showConfirmation("Fragment selection saved!");
    }
  }

  closeSaveSelectionModal() {
    document.getElementById("save-selection-modal").style.display = "none";
  }

  restoreSavedSelection(selection) {
    selection.forEach((index) => {
      this.store.dispatch(toggleCheckedFragment(index));
    });
    this.updateUI();
    showConfirmation("Saved selection restored!");
  }
}

export default FragmentsColumn;
