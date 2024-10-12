import { LitElement, html, css } from 'https://unpkg.com/lit@2.6.1/index.js?module';

class ImageList extends LitElement {
    static styles = css`
        #image-list {
            width: 150px;
            overflow-y: auto;
            border-right: 1px solid #ccc;
            padding: 10px;
        }
        #image-list img {
            width: 100%;
            cursor: pointer;
            margin-bottom: 10px;
            border: 2px solid transparent;
            transition: border 0.3s;
        }
        #image-list img.selected {
            border: 2px solid #007BFF;
        }
    `;

    static properties = {
        images: { type: Array },
        activeIndex: { type: Number },
    };

    render() {
        return html`
            <div id="image-list">
                ${this.images.map((image, index) => html`
                    <img 
                        src="${image.img.src}" 
                        data-index="${index}" 
                        class="${this.activeIndex === index ? 'selected' : ''}"
                        @click="${() => this.selectImage(index)}"
                        alt="Image ${index + 1}"
                    />
                `)}
            </div>
        `;
    }

    selectImage(index) {
        this.dispatchEvent(new CustomEvent('image-selected', {
            detail: index,
            bubbles: true,
            composed: true
        }));
    }
}

customElements.define('image-list', ImageList);
