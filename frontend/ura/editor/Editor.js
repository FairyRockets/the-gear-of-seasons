import Preview from "./Preview";
import { throws } from "assert";

/**
  @typedef Moment
  @type {object}
  @property {string} title
  @property {string} date
  @property {string} author
  @property {string} text
*/

export default class Editor {
  /**
   * 
   * @param {HTMLInputElement} title 
   * @param {HTMLInputElement} date 
   * @param {HTMLSelectElement} author 
   * @param {HTMLTextAreaElement} text 
   * @param {HTMLButtonElement} submit
   */
  constructor(title, date, author, text, submit) {
    this.title_ = title;
    this.date_ = date;
    this.author_ = author;
    this.text_ = text;
    this.submit_ = submit;
    this.onChangeEventListener_ = this.onChange_.bind(this);
    /** @type {Preview} */
    this.preview_ = null;
    /** @type {number|null} */
    this.changeId_ = null;
    this.executePreviewUpdater_ = this.executePreviewUpdate_.bind(this);
    this.onSaveEventListener_ = this.onSave_.bind(this);
  }
  /**
   * 
   * @param {Preview} preview 
   */
  init(preview) {
    this.preview_ = preview;
    this.text_.addEventListener('input', this.onChangeEventListener_);
    this.title_.addEventListener('input', this.onChangeEventListener_);
    this.date_.addEventListener('input', this.onChangeEventListener_);
    this.author_.addEventListener('change', this.onChangeEventListener_);
    this.submit_.addEventListener('click', this.onSaveEventListener_);
    window.addEventListener('keypress', (event) => {
      if (!(event.which === 115 && event.ctrlKey) && !(event.which === 19)) return true;
      window.setTimeout(this.onSaveEventListener_, 0);
      event.preventDefault();
      return false;
    });
    // reload対策
    if(this.text_.value.length > 0 || this.title_.value.length > 0) {
      this.onChange_();
    }
    this.submit_.disabled = true;
  }
  /**
   * @private
   */
  onChange_() {
    this.submit_.disabled = false;
    if(!!this.changeId_) {
      return;
    }
    this.changeId_ = setTimeout(this.executePreviewUpdater_, 500);
  }
  /**
   * @private
   */
  makeMoment_() {
    return {
      title: this.title_.value,
      date: this.date_.value.length > 0 ? this.date_.value : null,
      author: this.author_.value,
      text: this.text_.value
    };
  }
  /**
   * @private
   */
  onSave_() {
    this.changeId_ = null;
    const moment = this.makeMoment_();
    fetch('/save', {
      method: 'POST',
      body: JSON.stringify(moment),
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
      }}).then(result => result.json())
      .then(result => {
        this.submit_.disabled = true;
        this.preview_.onChange(result.body);
        history.replaceState(null, null, result.path)
      });
  }
  executePreviewUpdate_() {
    this.changeId_ = null;
    const moment = this.makeMoment_();
    fetch('/preview', {
      method: 'POST',
      body: JSON.stringify(moment),
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json'
      }}).then(result => result.text())
      .then(body => this.preview_.onChange(body));
  }
  /**
   * 
   * @param {string[]} embeds 
   */
  onUpload(embeds) {
    this.text_.value += '\n' + embeds.join('\n\n');
    this.executePreviewUpdate_();
  }
}

