let apps = [];

const sections = [
  { id: "overview", label: "Overview", symbol: "⊞" },
  { id: "artifacts", label: "Release artifacts", symbol: "◈" },
  { id: "releases", label: "Releases", symbol: "↗" },
  { id: "environments", label: "Environments", symbol: "⌘" },
  { id: "services", label: "Services", symbol: "◇" }
];

async function loadApps() {
  try {
    const response = await fetch('/apps');
    if (!response.ok) throw new Error('Failed to fetch apps');
    const data = await response.json();

    apps = data.map(app => ({
      id: app.app_id,
      name: app.name,
      mark: app.mark,
      description: app.description
    }));

    // Re-render components after data is loaded
    renderApps();
    renderDirectory();
    if (!showingDirectory) {
      renderContent();
    }
  } catch (error) {
    console.error('Error loading apps:', error);
  }
}

async function loadServices(appId) {
  try {
    const response = await fetch(`/services?app_id=${appId}`);
    if (!response.ok) throw new Error('Failed to fetch services');
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error loading services:', error);
    return [];
  }
}

const releaseRows = [
  ["v2.14.0", "Improve edge caching", "RS", "2m ago"],
  ["v2.13.2", "Add retry policy", "KM", "Yesterday"],
  ["v2.13.1", "Update dependencies", "RS", "3 days ago"]
];

const environmentRows = [
  ["Production", "v2.14.0 · 2m ago", "Live", "P"],
  ["Staging", "v2.14.0 · 11m ago", "Live", "S"],
  ["Development", "v2.15.0-rc.2 · 1h ago", "Live", "D"]
];

const artifactOptions = [
  { id: "artifact-1", label: "Artifact-1 · Improve edge caching" },
  { id: "artifact-2", label: "Artifact-2 · Add retry policy" },
  { id: "artifact-3", label: "Artifact-3 · Update dependencies" }
];

let environmentRecords = [
  { id: "production", name: "Production", url: "orbit.app", secret: "prod-secret", secretVisible: true, service: "Orbit API", artifact: "artifact-1", status: "Live", statusClass: "healthy-health" }
];

let selectedApp = "orbit";
let selectedSection = "overview";
let showingDirectory = true;
let environmentEditorOpen = false;
let editingEnvironmentId = null;
let environmentFormOpen = false;
let environmentChanges = [];
let preselectedArtifactId = null;
let releaseDetailOpen = false;
let selectedReleaseIndex = 0;

const appList = document.querySelector("#app-list");
const sectionNav = document.querySelector("#section-nav");
const appShell = document.querySelector(".app-shell");
const appsDirectory = document.querySelector("#apps-directory");
const appContent = document.querySelector("#app-content");
const directoryList = document.querySelector("#directory-list");
const directoryCount = document.querySelector("#directory-count");
const emptyDirectory = document.querySelector("#empty-directory");
const appSearch = document.querySelector("#app-search");
const sidebarContextName = document.querySelector("#sidebar-context-name");
const breadcrumbs = document.querySelector("#breadcrumbs");
const pageTitle = document.querySelector("#page-title");
const pageDescription = document.querySelector("#page-description");
const breadcrumbApp = document.querySelector("#breadcrumb-app");
const breadcrumbSection = document.querySelector("#breadcrumb-section");
const releaseTable = document.querySelector("#release-table");
const environmentList = document.querySelector("#environment-list");
const sectionContent = document.querySelector("#section-content");
const dashboardContent = document.querySelector("#dashboard-content");

function renderApps() {
  appList.innerHTML = apps.map((app) => `
    <button class="app-item ${!showingDirectory && app.id === selectedApp ? "active" : ""}" type="button" data-app="${app.id}" aria-pressed="${!showingDirectory && app.id === selectedApp}">
      <span class="app-logo">${app.mark}</span><span class="app-name">${app.name}</span><span class="app-status"></span>
    </button>
  `).join("");
}

function renderDirectory(query = "") {
  const normalizedQuery = query.trim().toLowerCase();
  const visibleApps = apps.filter((app) => `${app.name} ${app.description}`.toLowerCase().includes(normalizedQuery));
  directoryCount.textContent = `${visibleApps.length} ${visibleApps.length === 1 ? "app" : "apps"}`;
  emptyDirectory.hidden = visibleApps.length > 0;
  directoryList.innerHTML = visibleApps.map((app) => `
    <button class="directory-card" type="button" data-directory-app="${app.id}">
      <span class="directory-card-top"><span class="directory-logo">${app.mark}</span><span class="app-status"></span></span>
      <strong>${app.name}</strong><span>${app.description}</span><small>Open app <b>→</b></small>
    </button>
  `).join("");
}

function showDirectory() {
  showingDirectory = true;
  appShell.classList.add("directory-mode");
  appsDirectory.hidden = false;
  appContent.hidden = true;
  document.querySelector("#breadcrumb-app").textContent = "All apps";
  document.querySelector("#breadcrumb-section").textContent = "Directory";
  renderApps();
}

function showApp() {
  showingDirectory = false;
  appShell.classList.remove("directory-mode");
  appsDirectory.hidden = true;
  appContent.hidden = false;
  sidebarContextName.textContent = apps.find((app) => app.id === selectedApp).name;
  renderApps();
  renderSections();
  renderContent();
}

function renderSections() {
  sectionNav.innerHTML = sections.map((section) => `
    <button class="section-item ${section.id === selectedSection ? "active" : ""}" type="button" data-section="${section.id}" aria-current="${section.id === selectedSection ? "page" : "false"}">
      <span class="section-symbol">${section.symbol}</span><span>${section.label}</span>
    </button>
  `).join("");
}

function renderContent() {
  const app = apps.find((item) => item.id === selectedApp);
  const section = sections.find((item) => item.id === selectedSection);
  const editingEnvironment = selectedSection === "environments" && environmentEditorOpen;
  const showingReleaseDetail = selectedSection === "releases" && releaseDetailOpen;
  const selectedRelease = ["Improve edge caching", "Add retry policy", "Update dependencies"][selectedReleaseIndex];
  pageTitle.textContent = editingEnvironment ? (editingEnvironmentId ? "Configure environment" : "Add environment") : showingReleaseDetail ? selectedRelease : selectedSection === "overview" ? app.name : section.label;
  pageDescription.textContent = editingEnvironment ? "Set the environment target, applied service, and artifact attachment." : showingReleaseDetail ? "Deployment details for the selected production release." : selectedSection === "overview" ? app.description : `${section.label} for ${app.name}. Track the work and keep your team moving.`;
  breadcrumbApp.textContent = app.name;
  breadcrumbSection.textContent = editingEnvironment ? (editingEnvironmentId ? "Configure" : "New environment") : showingReleaseDetail ? "Release details" : section.label;
  appContent.dataset.section = selectedSection;
  if (editingEnvironment) renderEnvironmentEditor();
  else if (showingReleaseDetail) renderReleaseDetail();
  else renderSectionContent();
}

function renderReleaseDetail() {
  const release = [
    { commit: "Improve edge caching", status: "Active", deployed: "2m ago", by: "RS" },
    { commit: "Add retry policy", status: "Completed", deployed: "Yesterday", by: "KM" },
    { commit: "Update dependencies", status: "Completed", deployed: "3 days ago", by: "RS" }
  ][selectedReleaseIndex];
  sectionContent.hidden = false;
  dashboardContent.hidden = true;
  sectionContent.innerHTML = `<section class="release-detail-page"><button class="text-button back-release" type="button" data-release-action="back-to-releases">← Back to releases</button><div class="release-detail-header"><div class="release-commit-heading"><span class="user-avatar">${release.by}</span><div><p class="eyebrow">Production release</p><h3>${release.commit}</h3></div></div><span class="release-status">${release.status}</span></div><div class="release-detail-meta"><span>Production</span><span>${release.deployed}</span></div><div class="release-detail-steps"><div class="release-step complete"><span>01</span><div><strong>Artifact selected</strong><small>${release.commit}</small></div></div><div class="release-step complete"><span>02</span><div><strong>Deployed to Production</strong><small>${release.deployed}</small></div></div><div class="release-step ${release.status === "Active" ? "complete" : "pending"}"><span>03</span><div><strong>${release.status === "Active" ? "Health checks passed" : "Rollback available"}</strong><small>${release.status === "Active" ? "All services healthy" : "Ready to restore"}</small></div></div></div></section>`;
}

function renderEnvironmentEditor() {
  const environment = environmentRecords.find((item) => item.id === editingEnvironmentId);
  sectionContent.hidden = false;
  dashboardContent.hidden = true;
  sectionContent.innerHTML = `
    <div class="environment-deploy-page">
      <section class="select-artifact-card"><p class="eyebrow">Release target</p><h3>Select Artifact</h3><select id="deploy-artifact-select" aria-label="Select artifact to deploy">${artifactOptions.map((artifact) => `<option value="${artifact.id}">${artifact.label}</option>`).join("")}</select></section>
      ${environmentFormOpen ? `<section class="environment-inline-form"><div class="editor-card-header"><div><p class="eyebrow">${environment ? "Update environment" : "New environment"}</p><h3>${environment ? environment.name : "Add environment"}</h3></div><span class="editor-step">${environment ? "Edit" : "Create"}</span></div><form id="environment-management-form" class="environment-form"><label><span>ENV NAME</span><input name="name" required value="${environment?.name || ""}" placeholder="Production" /></label><label><span>SECRET</span><input name="secret" type="password" required value="${environment?.secret || ""}" placeholder="Environment secret" /></label><label><span>SERVICE NAME</span><select name="service" required><option>All services</option><option>Orbit API</option><option>Worker queue</option><option>Image processor</option></select></label><div class="editor-actions"><button class="secondary-button" type="button" data-env-action="cancel-form">Cancel</button><button class="primary-button" type="submit">${environment ? "Update environment" : "Add environment"}</button></div></form></section>` : ""}
      <section class="environment-manage-card"><div class="environment-manage-header"><p class="eyebrow">Environment targets</p><span>${environmentRecords.length} environments</span></div><div class="environment-list-header"><span>ENV NAME</span><span>SECRET</span><span>SERVICE NAME</span><span>ACTIONS</span></div>
        <div class="environment-manage-list">${environmentRecords.map((item) => `
          <div class="environment-manage-row"><strong>${item.name}</strong><span>${item.secretVisible ? item.secret : "••••••••"}</span><span>${item.service}</span><span class="row-actions"><button type="button" data-env-action="configure" data-environment-id="${item.id}">Edit</button><button type="button" data-env-action="delete" data-environment-id="${item.id}">Delete</button></span></div>
        `).join("")}</div>
      </section>
      ${environmentChanges.length ? `<section class="environment-changes-card"><div class="environment-manage-header"><p class="eyebrow">Recent changes</p><span>${environmentChanges.length} pending</span></div><div class="environment-change-list">${environmentChanges.map((change) => `<div class="environment-change-row"><span class="change-indicator">${change.action === "deleted" ? "−" : change.action === "created" ? "+" : "✎"}</span><div><strong>${change.label}</strong><small>${change.action === "deleted" ? "Environment deleted" : change.action === "created" ? "Environment created" : "Environment updated"}</small></div><button class="secondary-button" type="button" data-env-action="undo" data-change-id="${change.id}">Undo</button></div>`).join("")}</div></section>` : ""}
      ${environmentFormOpen ? "" : `<button class="add-environment-button" type="button" data-env-action="new-row">+ Add New ENV</button>`}
      <div class="deploy-page-footer"><span class="attachment-note"><span>↗</span><span>Selected artifact will redeploy to the applied services in these environments.</span></span><button class="primary-button" type="button" data-env-action="deploy">Deploy</button></div>
    </div>`;
  const deployArtifactSelect = document.querySelector("#deploy-artifact-select");
  if (environment) deployArtifactSelect.value = environment.artifact;
  else if (preselectedArtifactId) deployArtifactSelect.value = preselectedArtifactId;
  if (environment) document.querySelector("#environment-management-form").service.value = environment.service;
  const managementForm = document.querySelector("#environment-management-form");
  const secretInput = document.querySelector("#environment-management-form input[name=\"secret\"]");
  if (managementForm && secretInput) {
    const secretToggle = document.createElement("span");
    secretToggle.className = "secret-toggle";
    secretToggle.innerHTML = `<input type="checkbox" aria-label="Show secret" /><span>Show secret</span>`;
    secretInput.insertAdjacentElement("afterend", secretToggle);
    secretToggle.querySelector("input").addEventListener("change", (event) => {
      secretInput.type = event.target.checked ? "text" : "password";
    });
  }
  if (managementForm) managementForm.addEventListener("submit", (event) => {
    event.preventDefault();
    const values = Object.fromEntries(new FormData(managementForm));
    const previous = editingEnvironmentId ? environmentRecords.find((item) => item.id === editingEnvironmentId) : null;
    const record = { id: editingEnvironmentId || values.name.toLowerCase().replace(/\s+/g, "-"), name: values.name, url: previous?.url || `${values.name.toLowerCase().replace(/\s+/g, "-")}.orbit.app`, secret: values.secret, secretVisible: secretInput.type === "text", service: values.service, artifact: document.querySelector("#deploy-artifact-select").value, status: "Live", statusClass: "healthy-health" };
    if (editingEnvironmentId) {
      environmentChanges.unshift({ id: `change-${Date.now()}`, action: "updated", label: record.name, before: previous });
      environmentRecords = environmentRecords.map((item) => item.id === editingEnvironmentId ? record : item);
    } else {
      environmentRecords.push(record);
      environmentChanges.unshift({ id: `change-${Date.now()}`, action: "created", label: record.name, created: record });
    }
    environmentFormOpen = false;
    editingEnvironmentId = null;
    renderEnvironmentEditor();
  });
}

async function renderSectionContent() {
  if (selectedSection === "overview") {
    dashboardContent.hidden = false;
    sectionContent.hidden = true;
    return;
  }

  dashboardContent.hidden = true;
  sectionContent.hidden = false;
  if (selectedSection === "artifacts") {
    sectionContent.innerHTML = `
      <section class="release-artifact-stack">
        <div class="service-stack-header">
          <div><p class="eyebrow">Container artifacts</p><h3>Latest tagged versions</h3></div>
        </div>
        <div id="release-artifacts-content"><p class="empty-directory">Loading latest artifacts...</p></div>
      </section>`;

    try {
      const response = await fetch("/getPackages");
      if (!response.ok) throw new Error("Failed to fetch release artifacts");

      const data = await response.json();
      const releases = Array.isArray(data.releases) ? data.releases : [];
      const container = document.querySelector("#release-artifacts-content");

      if (!releases.length) {
        container.innerHTML = '<p class="empty-directory">No tagged container artifacts found.</p>';
        return;
      }

      container.innerHTML = releases.map((release, index) => {
        const packages = Array.isArray(release.packages) ? release.packages : [];
        return `
          <section class="release-artifact-box">
            <div class="release-artifact-header">
              <div><h3>${release.tag}</h3><span>Latest tag for ${packages.length === 1 ? "this service" : "these services"}</span></div>
            </div>
            <div class="release-artifact-list">
              ${packages.map((pkg) => `
                <div class="release-artifact-row">
                  <span class="docker-mark">▣</span>
                  <div><strong>Docker image</strong><span>${pkg.packageName}</span></div>
                </div>
              `).join("")}
            </div>
            <div class="release-artifact-footer">
              <span>${packages.length} service artifact${packages.length === 1 ? "" : "s"}</span>
              <div class="artifact-actions">
                <button class="secondary-button" type="button" data-release-action="draft" data-artifact-id="artifact-${index}">Draft</button>
                <button class="primary-button" type="button" data-release-action="release-now">Release now</button>
              </div>
            </div>
          </section>`;
      }).join("");
    } catch (error) {
      console.error("Error loading release artifacts:", error);
      document.querySelector("#release-artifacts-content").innerHTML =
        '<p class="empty-directory">Unable to load release artifacts.</p>';
    }
  }
  if (selectedSection === "releases") {
    const releaseCards = [
      { version: "v2.14.0", commit: "Improve edge caching", status: "Active", statusClass: "active-deploy", deployed: "2m ago", by: "RS" },
      { version: "v2.13.2", commit: "Add retry policy", status: "Completed", statusClass: "completed-deploy", deployed: "Yesterday", by: "KM" },
      { version: "v2.13.1", commit: "Update dependencies", status: "Completed", statusClass: "completed-deploy", deployed: "3 days ago", by: "RS" }
    ];
    sectionContent.innerHTML = `
      <section class="release-card-stack"><div class="service-stack-header"><div><p class="eyebrow">Production delivery</p><h3>Deployments</h3></div></div>
          ${releaseCards.map((release, index) => `
            <article class="release-card ${release.statusClass}" data-release-index="${index}">
              <div class="release-card-header"><div class="release-commit-heading"><span class="user-avatar">${release.by}</span><div><h3>${release.commit}</h3><p>Production · ${release.deployed}</p></div></div><div class="release-status-stack"><span class="release-status">${release.status}</span><button class="text-button show-release" type="button" data-release-action="show-release">Show release</button></div></div>
              <div class="release-card-details"><div><span class="eyebrow">Environment</span><strong>Production</strong></div><button class="secondary-button release-action" type="button" data-release-action="${release.status === "Active" ? "restart" : "rollback"}">${release.status === "Active" ? "Restart" : "Rollback"}</button></div>
              <div class="release-steps" hidden><div class="release-step complete"><span>01</span><div><strong>Artifact selected</strong><small>${release.commit}</small></div></div><div class="release-step complete"><span>02</span><div><strong>Deployed to Production</strong><small>${release.deployed}</small></div></div><div class="release-step ${release.status === "Active" ? "complete" : "pending"}"><span>03</span><div><strong>${release.status === "Active" ? "Health checks passed" : "Rollback available"}</strong><small>${release.status === "Active" ? "All services healthy" : "Ready to restore"}</small></div></div></div>
            </article>
          `).join("")}
      </section>`;
  }
  if (selectedSection === "environments") {
    sectionContent.innerHTML = `
      <section class="environment-card-stack"><div class="service-stack-header"><div><p class="eyebrow">Runtime targets</p><h3>Environments</h3></div><button class="primary-button" type="button" data-env-action="add"><span>+</span> Add environment</button></div>
        <div class="environment-toolbar"><label class="search-field" for="environment-search"><span>⌕</span><input id="environment-search" type="search" placeholder="Search environments" autocomplete="off" /></label><label class="service-filter" for="environment-service-filter"><span>Service</span><select id="environment-service-filter"><option value="all">All services</option><option value="Orbit API">Orbit API</option><option value="Worker queue">Worker queue</option><option value="Image processor">Image processor</option></select></label></div>
        <p class="empty-directory" id="environment-empty" hidden>No environments match your filters.</p>
          ${environmentRecords.map((environment) => `
            <article class="environment-card" data-environment-name="${environment.name.toLowerCase()}" data-environment-url="${environment.url.toLowerCase()}" data-environment-service="${environment.service}">
              <div class="environment-card-header"><div><h3>${environment.name}</h3><p>Secret · ${environment.secretVisible ? environment.secret : "••••••••"}</p></div><span class="service-health ${environment.statusClass}">${environment.status}</span></div>
              <div class="environment-card-details"><div><span class="eyebrow">Applied service</span><strong>${environment.service}</strong></div><div><span class="eyebrow">Deployments</span><strong>Automatic</strong></div></div>
            </article>
          `).join("")}
      </section>`;
    const environmentSearch = document.querySelector("#environment-search");
    const environmentServiceFilter = document.querySelector("#environment-service-filter");
    const environmentEmpty = document.querySelector("#environment-empty");
    const filterEnvironments = () => {
      const query = environmentSearch.value.trim().toLowerCase();
      const service = environmentServiceFilter.value;
      let visibleCount = 0;
      document.querySelectorAll(".environment-card").forEach((card) => {
        const matchesQuery = `${card.dataset.environmentName} ${card.dataset.environmentUrl} ${card.dataset.environmentService.toLowerCase()}`.includes(query);
        const matchesService = service === "all" || card.dataset.environmentService === service;
        card.hidden = !(matchesQuery && matchesService);
        if (!card.hidden) visibleCount += 1;
      });
      environmentEmpty.hidden = visibleCount > 0;
    };
    environmentSearch.addEventListener("input", filterEnvironments);
    environmentServiceFilter.addEventListener("change", filterEnvironments);
  }
  if (selectedSection === "services") {
    const services = await loadServices(selectedApp);
    sectionContent.innerHTML = `
      <section class="service-card-stack"><div class="service-stack-header"><div><p class="eyebrow">Runtime inventory</p><h3>Active services</h3></div><button class="secondary-button" type="button">Service settings <span>↗</span></button></div>
        ${services.map((service) => `
          <article class="service-card ${service.health === "Degraded" ? "degraded-service" : ""}">
            <div class="service-card-header"><div><h3>${service.service_id}</h3><p>${service.runtime} · ${service.port}</p></div><span class="service-health ${service.health === "Healthy" ? "healthy-health" : "degraded-health"}">${service.health || "Healthy"}</span></div>
            <div class="service-card-details"><div><span class="eyebrow">Capacity</span><strong>${service.capacity || "N/A"}</strong></div><div><span class="eyebrow">Last deployed</span><strong>${service.deployed_at ? new Date(service.deployed_at).toLocaleString() : "N/A"}</strong></div><div class="service-card-actions"><button class="secondary-button" type="button" data-service-action="restart">Restart</button><button class="primary-button" type="button" data-service-action="scale">Scale</button></div></div>
          </article>
        `).join("")}
      </section>`;
  }
}

function renderData() {
  releaseTable.innerHTML = releaseRows.map(([version, title, user, date]) => `
    <div class="release-row"><div class="release-name">${version}<span>${title}</span></div><div class="release-user">${user}</div><div class="release-date">${date}</div><div class="release-state">Success</div></div>
  `).join("");
  environmentList.innerHTML = environmentRows.map(([name, detail, state, mark]) => `
    <div class="environment-row"><span class="environment-icon">${mark}</span><div class="environment-copy"><strong>${name}</strong><span>${detail}</span></div><span class="environment-status">${state}</span></div>
  `).join("");
}

sectionContent.addEventListener("click", (event) => {
  const button = event.target.closest("[data-env-action]");
  if (!button) return;
  const action = button.dataset.envAction;
  if (action === "add") {
    editingEnvironmentId = null;
    environmentEditorOpen = true;
    environmentFormOpen = false;
    renderContent();
  }
  if (action === "new-row") {
    editingEnvironmentId = null;
    environmentFormOpen = true;
    renderEnvironmentEditor();
  }
  if (action === "configure") {
    editingEnvironmentId = button.dataset.environmentId;
    environmentEditorOpen = true;
    environmentFormOpen = true;
    renderContent();
  }
  if (action === "delete") {
    const deleted = environmentRecords.find((item) => item.id === button.dataset.environmentId);
    if (deleted) environmentChanges.unshift({ id: `change-${Date.now()}`, action: "deleted", label: deleted.name, before: deleted });
    environmentRecords = environmentRecords.filter((item) => item.id !== button.dataset.environmentId);
    renderEnvironmentEditor();
  }
  if (action === "undo") {
    const change = environmentChanges.find((item) => item.id === button.dataset.changeId);
    if (!change) return;
    if (change.action === "deleted") environmentRecords.push(change.before);
    if (change.action === "created") environmentRecords = environmentRecords.filter((item) => item.id !== change.created.id);
    if (change.action === "updated") environmentRecords = environmentRecords.map((item) => item.id === change.before.id ? change.before : item);
    environmentChanges = environmentChanges.filter((item) => item.id !== change.id);
    renderEnvironmentEditor();
  }
  if (action === "cancel") {
    editingEnvironmentId = null;
    environmentEditorOpen = false;
    environmentFormOpen = false;
    renderContent();
  }
  if (action === "cancel-form") {
    editingEnvironmentId = null;
    environmentFormOpen = false;
    renderEnvironmentEditor();
  }
  if (action === "deploy") {
    environmentChanges = [];
    renderEnvironmentEditor();
  }
});

sectionContent.addEventListener("click", (event) => {
  const draftButton = event.target.closest("[data-release-action=\"draft\"]");
  if (draftButton) {
    preselectedArtifactId = draftButton.dataset.artifactId;
    selectedSection = "environments";
    environmentEditorOpen = true;
    environmentFormOpen = false;
    editingEnvironmentId = null;
    renderSections();
    renderContent();
    return;
  }
  const button = event.target.closest("[data-release-action=\"show-release\"]");
  if (!button) return;
  const card = button.closest(".release-card");
  selectedReleaseIndex = Number(card.dataset.releaseIndex);
  releaseDetailOpen = true;
  renderContent();
});

sectionContent.addEventListener("click", (event) => {
  const button = event.target.closest("[data-release-action=\"back-to-releases\"]");
  if (!button) return;
  releaseDetailOpen = false;
  renderContent();
});

appList.addEventListener("click", (event) => {
  const button = event.target.closest("[data-app]");
  if (!button) return;
  selectedApp = button.dataset.app;
  selectedSection = "overview";
  showApp();
});

directoryList.addEventListener("click", (event) => {
  const button = event.target.closest("[data-directory-app]");
  if (!button) return;
  selectedApp = button.dataset.directoryApp;
  selectedSection = "overview";
  showApp();
});

sectionNav.addEventListener("click", (event) => {
  const button = event.target.closest("[data-section]");
  if (!button) return;
  selectedSection = button.dataset.section;
  releaseDetailOpen = false;
  environmentEditorOpen = false;
  environmentFormOpen = false;
  editingEnvironmentId = null;
  renderSections();
  renderContent();
});

document.querySelectorAll("[data-global]").forEach((button) => {
  button.addEventListener("click", () => {
    document.querySelectorAll("[data-global]").forEach((item) => item.classList.remove("active"));
    button.classList.add("active");
    if (button.dataset.global === "apps") showDirectory();
  });
});

appSearch.addEventListener("input", (event) => renderDirectory(event.target.value));

breadcrumbs.addEventListener("click", (event) => {
  const button = event.target.closest("[data-breadcrumb]");
  if (!button) return;
  if (button.dataset.breadcrumb === "apps" || (button.dataset.breadcrumb === "app" && showingDirectory)) {
    showDirectory();
    return;
  }
  if (button.dataset.breadcrumb === "app") {
    selectedSection = "overview";
    environmentEditorOpen = false;
    editingEnvironmentId = null;
    showApp();
    return;
  }
  if (button.dataset.breadcrumb === "section" && environmentEditorOpen) {
    environmentEditorOpen = false;
    editingEnvironmentId = null;
  }
  if (button.dataset.breadcrumb === "section" && releaseDetailOpen) releaseDetailOpen = false;
  showApp();
});

document.querySelector(".release-panel").addEventListener("click", (event) => {
  if (event.target.closest("[data-section]")) {
    selectedSection = "releases";
    renderSections();
    renderContent();
  }
});

// Existing logic...
// Remove the immediate renders that use empty arrays
// renderApps();
// renderSections();
// renderContent();
// renderData();
// renderDirectory();
// showDirectory();

// Initialize dynamic data
loadApps();

// we still need these to initialize the UI structure
renderSections();
renderData();
showDirectory();


