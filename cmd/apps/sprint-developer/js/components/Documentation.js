import { html } from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';
import { BootstrapLitElement } from '../base-component.js';
import { processDocumentation } from '../documentation.js';

export class Documentation extends BootstrapLitElement {
  constructor() {
    super();
  }

  renderIntroduction() {
    return html`
      <div class="card mb-4">
        <div class="card-header">
          <h2 class="h4 mb-0">${processDocumentation.introduction.title}</h2>
        </div>
        <div class="card-body">
          <p>${processDocumentation.introduction.description}</p>
        </div>
      </div>
    `;
  }

  renderPreparation() {
    return html`
      <div class="card mb-4">
        <div class="card-header">
          <h2 class="h4 mb-0">${processDocumentation.preparation.title}</h2>
        </div>
        <div class="card-body">
          ${processDocumentation.preparation.steps.map(step => html`
            <div class="mb-3">
              <h4 class="h5">${step.title} (${step.ratio})</h4>
              <div class="ms-3">
                <p class="mb-1">${step.example.concentrate}</p>
                <p class="mb-1">${step.example.water}</p>
                ${step.example.hardener ? html`<p class="mb-1">${step.example.hardener}</p>` : ''}
                <p class="mb-1 fw-bold">${step.example.result}</p>
              </div>
            </div>
          `)}
        </div>
      </div>
    `;
  }

  renderProcessSteps() {
    return html`
      <div class="card mb-4">
        <div class="card-header">
          <h2 class="h4 mb-0">${processDocumentation.process.title}</h2>
        </div>
        <div class="card-body p-0">
          <div class="accordion" id="processSteps">
            ${processDocumentation.process.steps.map(step => html`
              <div class="accordion-item">
                <h3 class="accordion-header">
                  <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#step${step.id}">
                    ${step.id}. ${step.name} (${step.duration})
                  </button>
                </h3>
                <div id="step${step.id}" class="accordion-collapse collapse">
                  <div class="accordion-body">
                    <dl class="row mb-0">
                      <dt class="col-sm-3">Purpose</dt>
                      <dd class="col-sm-9">${step.details.purpose}</dd>
                      
                      <dt class="col-sm-3">Instructions</dt>
                      <dd class="col-sm-9">${step.details.instructions}</dd>
                      
                      ${step.details.agitation ? html`
                        <dt class="col-sm-3">Agitation</dt>
                        <dd class="col-sm-9">${step.details.agitation}</dd>
                      ` : ''}
                      
                      <dt class="col-sm-3">Notes</dt>
                      <dd class="col-sm-9">${step.details.notes}</dd>
                      
                      ${step.details.warnings ? html`
                        <dt class="col-sm-3 text-danger">Warning</dt>
                        <dd class="col-sm-9 text-danger">${step.details.warnings}</dd>
                      ` : ''}
                    </dl>
                  </div>
                </div>
              </div>
            `)}
          </div>
        </div>
      </div>
    `;
  }

  renderTemperatureChart() {
    return html`
      <div class="card mb-4">
        <div class="card-header">
          <h2 class="h4 mb-0">${processDocumentation.temperatureChart.title}</h2>
        </div>
        <div class="card-body">
          <p>${processDocumentation.temperatureChart.description}</p>
          <div class="table-responsive">
            <table class="table table-bordered">
              <thead>
                <tr>
                  <th>Celsius</th>
                  <th>Fahrenheit</th>
                </tr>
              </thead>
              <tbody>
                ${processDocumentation.temperatureChart.temperatures.map(temp => html`
                  <tr>
                    <td>${temp.celsius}°C</td>
                    <td>${temp.fahrenheit}°F</td>
                  </tr>
                `)}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    `;
  }

  renderSafetyGuidelines() {
    return html`
      <div class="card mb-4">
        <div class="card-header">
          <h2 class="h4 mb-0">${processDocumentation.safetyGuidelines.title}</h2>
        </div>
        <div class="card-body">
          <h3 class="h5">General Guidelines</h3>
          <ul class="mb-4">
            ${processDocumentation.safetyGuidelines.general.map(guideline => html`
              <li>${guideline}</li>
            `)}
          </ul>
          
          <h3 class="h5">First Aid</h3>
          <dl class="row mb-0">
            <dt class="col-sm-3">Eye Contact</dt>
            <dd class="col-sm-9">${processDocumentation.safetyGuidelines.firstAid.eyeContact}</dd>
            
            <dt class="col-sm-3">Skin Contact</dt>
            <dd class="col-sm-9">${processDocumentation.safetyGuidelines.firstAid.skinContact}</dd>
            
            <dt class="col-sm-3">Inhalation</dt>
            <dd class="col-sm-9">${processDocumentation.safetyGuidelines.firstAid.inhalation}</dd>
            
            <dt class="col-sm-3">Ingestion</dt>
            <dd class="col-sm-9">${processDocumentation.safetyGuidelines.firstAid.ingestion}</dd>
          </dl>
        </div>
      </div>
    `;
  }

  render() {
    return html`
      <div class="row">
        <div class="col-12">
          ${this.renderIntroduction()}
          ${this.renderPreparation()}
          ${this.renderProcessSteps()}
          ${this.renderTemperatureChart()}
          ${this.renderSafetyGuidelines()}
        </div>
      </div>
    `;
  }
} 