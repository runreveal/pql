import { Application, Controller } from 'https://unpkg.com/@hotwired/stimulus/dist/stimulus.js';

window.Stimulus = Application.start();
Stimulus.register('analysis', class extends Controller {
  static targets = ['list', 'editor', 'output'];
  static values = {
    compileHref: String,
    suggestHref: String,
  };

  /**
   * @param {Event} event
   * @return {Promise<void>}
   */
  async compile(event) {
    const formData = new URLSearchParams();
    formData.append('source', this.editorTarget.value);
    const response = await fetch(this.compileHrefValue, {
      method: 'POST',
      body: formData,
    });
    if (!response.ok) {
      return;
    }
    this.outputTarget.innerHTML = await response.text();
    this.outputTarget.hidden = false;
  }

  /**
   * @param {Event} event
   * @return {Promise<void>}
   */
  async suggest(event) {
    const formData = new URLSearchParams();
    formData.append('source', this.editorTarget.value);
    formData.append('start', this.editorTarget.selectionStart);
    formData.append('end', this.editorTarget.selectionEnd);
    const response = await fetch(this.suggestHrefValue, {
      method: 'POST',
      body: formData,
    });
    if (!response.ok) {
      return;
    }
    this.listTarget.innerHTML = await response.text();
    this.listTarget.hidden = false;
  }

  clear() {
    this.listTarget.innerHTML = '';
    this.listTarget.hidden = true;
    this.editorTarget.focus();
  }

  /**
   * @param {Event} event
   */
  fill(event) {
    const i = this.editorTarget.selectionEnd;
    this.editorTarget.value = this.editorTarget.value.substring(0, i) +
      event.params.insert +
      this.editorTarget.value.substring(i);
    const insertEnd = i + event.params.insert.length;
    this.editorTarget.setSelectionRange(insertEnd, insertEnd);

    this.clear();
  }
});
