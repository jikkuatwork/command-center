// Command Center - Main Application JS

// State management
const state = {
  currentPage: 'dashboard',
  theme: localStorage.getItem('cc-theme') || 'light',
  stats: null,
  events: [],
  redirects: [],
  webhooks: [],
  domains: [],
  tags: []
};

// API client
const api = {
  async get(endpoint) {
    const response = await fetch(`/api/${endpoint}`);
    if (!response.ok) throw new Error(`API error: ${response.statusText}`);
    return await response.json();
  },

  async post(endpoint, data) {
    const response = await fetch(`/api/${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    });
    if (!response.ok) throw new Error(`API error: ${response.statusText}`);
    return await response.json();
  }
};

// Router
function navigate(page) {
  state.currentPage = page;
  document.getElementById('page-title').textContent =
    page.charAt(0).toUpperCase() + page.slice(1);

  // Update active nav item
  document.querySelectorAll('.nav-link').forEach(link => {
    link.classList.remove('active');
    if (link.dataset.page === page) {
      link.classList.add('active');
    }
  });

  // Load page content
  loadPage(page);
}

// Load page content
async function loadPage(page) {
  const content = document.getElementById('app-content');
  content.innerHTML = '<div class="text-center py-5"><div class="loading-spinner" style="width: 3rem; height: 3rem; border-width: 4px;"></div></div>';

  try {
    switch (page) {
      case 'dashboard':
        await loadDashboard();
        break;
      case 'analytics':
        await loadAnalytics();
        break;
      case 'redirects':
        await loadRedirects();
        break;
      case 'webhooks':
        await loadWebhooks();
        break;
      case 'settings':
        await loadSettings();
        break;
      default:
        content.innerHTML = '<div class="alert alert-danger">Page not found</div>';
    }
  } catch (error) {
    console.error('Error loading page:', error);
    content.innerHTML = `<div class="alert alert-danger">Error loading page: ${error.message}</div>`;
  }
}

// Load dashboard
async function loadDashboard() {
  const content = document.getElementById('app-content');

  // Fetch data
  state.stats = await api.get('stats');
  state.events = await api.get('events?limit=20');

  // Render dashboard
  content.innerHTML = `
    <!-- Stats Cards -->
    <div class="row row-deck row-cards mb-3">
      <div class="col-sm-6 col-lg-3">
        <div class="card stats-card">
          <div class="card-body">
            <div class="d-flex align-items-center">
              <div class="subheader">Events Today</div>
            </div>
            <div class="h1 mb-1">${state.stats.total_events_today.toLocaleString()}</div>
            <div class="text-muted">
              <i class="ti ti-trending-up"></i>
              ${state.stats.total_events_week.toLocaleString()} this week
            </div>
          </div>
        </div>
      </div>
      <div class="col-sm-6 col-lg-3">
        <div class="card stats-card">
          <div class="card-body">
            <div class="d-flex align-items-center">
              <div class="subheader">Total Events</div>
            </div>
            <div class="h1 mb-1">${state.stats.total_events_all_time.toLocaleString()}</div>
            <div class="text-muted">
              ${state.stats.total_events_month.toLocaleString()} this month
            </div>
          </div>
        </div>
      </div>
      <div class="col-sm-6 col-lg-3">
        <div class="card stats-card">
          <div class="card-body">
            <div class="d-flex align-items-center">
              <div class="subheader">Unique Domains</div>
            </div>
            <div class="h1 mb-1">${state.stats.total_unique_domains.toLocaleString()}</div>
            <div class="text-muted">
              ${state.stats.top_domains.length} active
            </div>
          </div>
        </div>
      </div>
      <div class="col-sm-6 col-lg-3">
        <div class="card stats-card">
          <div class="card-body">
            <div class="d-flex align-items-center">
              <div class="subheader">Redirect Clicks</div>
            </div>
            <div class="h1 mb-1">${state.stats.total_redirect_clicks.toLocaleString()}</div>
            <div class="text-muted">
              All time total
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Charts Row -->
    <div class="row row-cards mb-3">
      <div class="col-lg-8">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Traffic Timeline (Last 24 Hours)</h3>
          </div>
          <div class="card-body">
            <div class="chart-container">
              <canvas id="traffic-chart"></canvas>
            </div>
          </div>
        </div>
      </div>
      <div class="col-lg-4">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Events by Source</h3>
          </div>
          <div class="card-body">
            <div class="chart-container">
              <canvas id="source-chart"></canvas>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Top Domains & Recent Events -->
    <div class="row row-cards">
      <div class="col-lg-6">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Top Domains</h3>
          </div>
          <div class="card-table table-responsive">
            <table class="table table-vcenter">
              <thead>
                <tr>
                  <th>Domain</th>
                  <th class="text-end">Events</th>
                </tr>
              </thead>
              <tbody>
                ${state.stats.top_domains.slice(0, 10).map(d => `
                  <tr>
                    <td class="table-truncate">${d.domain}</td>
                    <td class="text-end">${d.count.toLocaleString()}</td>
                  </tr>
                `).join('')}
              </tbody>
            </table>
          </div>
        </div>
      </div>
      <div class="col-lg-6">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Recent Events</h3>
          </div>
          <div class="card-table table-responsive">
            <table class="table table-vcenter">
              <thead>
                <tr>
                  <th>Domain</th>
                  <th>Type</th>
                  <th>Time</th>
                </tr>
              </thead>
              <tbody>
                ${state.events.slice(0, 10).map(e => `
                  <tr>
                    <td class="table-truncate">${e.domain}</td>
                    <td><span class="badge">${e.event_type}</span></td>
                    <td class="text-muted">${timeAgo(e.created_at)}</td>
                  </tr>
                `).join('')}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  `;

  // Render charts
  renderTrafficChart();
  renderSourceChart();
}

// Render traffic timeline chart
function renderTrafficChart() {
  const ctx = document.getElementById('traffic-chart');
  if (!ctx) return;

  const timeline = state.stats.events_timeline || [];
  const labels = timeline.map(t => {
    const date = new Date(t.timestamp);
    return date.toLocaleTimeString('en-US', { hour: 'numeric', hour12: true });
  });
  const data = timeline.map(t => t.count);

  new Chart(ctx, {
    type: 'line',
    data: {
      labels: labels,
      datasets: [{
        label: 'Events',
        data: data,
        borderColor: '#206bc4',
        backgroundColor: 'rgba(32, 107, 196, 0.1)',
        tension: 0.4,
        fill: true
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { display: false }
      },
      scales: {
        y: { beginAtZero: true }
      }
    }
  });
}

// Render source type pie chart
function renderSourceChart() {
  const ctx = document.getElementById('source-chart');
  if (!ctx) return;

  const sourceTypes = state.stats.events_by_source_type || {};
  const labels = Object.keys(sourceTypes);
  const data = Object.values(sourceTypes);

  new Chart(ctx, {
    type: 'doughnut',
    data: {
      labels: labels,
      datasets: [{
        data: data,
        backgroundColor: [
          '#206bc4',
          '#2fb344',
          '#f59f00',
          '#d63939'
        ]
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { position: 'bottom' }
      }
    }
  });
}

// Load Analytics page with filtering
async function loadAnalytics() {
  const content = document.getElementById('app-content');

  // Fetch data
  state.events = await api.get('events?limit=100');
  state.domains = await api.get('domains');
  state.tags = await api.get('tags');

  content.innerHTML = `
    <div class="row mb-3">
      <div class="col">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Filters</h3>
          </div>
          <div class="card-body">
            <div class="row g-2">
              <div class="col-md-3">
                <select class="form-select" id="filter-domain">
                  <option value="">All Domains</option>
                  ${state.domains.map(d => `<option value="${d.domain}">${d.domain}</option>`).join('')}
                </select>
              </div>
              <div class="col-md-3">
                <select class="form-select" id="filter-source">
                  <option value="">All Sources</option>
                  <option value="web">Web</option>
                  <option value="pixel">Pixel</option>
                  <option value="redirect">Redirect</option>
                  <option value="webhook">Webhook</option>
                </select>
              </div>
              <div class="col-md-4">
                <input type="text" class="form-control" id="filter-search" placeholder="Search events...">
              </div>
              <div class="col-md-2">
                <button class="btn btn-primary w-100" onclick="applyFilters()">Apply</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-header">
        <h3 class="card-title">Events</h3>
        <div class="ms-auto">
          <span class="badge bg-blue">${state.events.length} events</span>
        </div>
      </div>
      <div class="card-table table-responsive">
        <table class="table table-vcenter">
          <thead>
            <tr>
              <th>Time</th>
              <th>Domain</th>
              <th>Type</th>
              <th>Source</th>
              <th>Path</th>
              <th>IP Address</th>
            </tr>
          </thead>
          <tbody id="events-table">
            ${renderEventsTable(state.events)}
          </tbody>
        </table>
      </div>
    </div>
  `;
}

function renderEventsTable(events) {
  return events.map(e => `
    <tr>
      <td class="text-muted">${timeAgo(e.created_at)}</td>
      <td class="table-truncate">${e.domain}</td>
      <td><span class="badge">${e.event_type}</span></td>
      <td><span class="badge bg-secondary">${e.source_type}</span></td>
      <td class="table-truncate">${e.path || '-'}</td>
      <td class="text-muted">${e.ip_address}</td>
    </tr>
  `).join('');
}

async function applyFilters() {
  const domain = document.getElementById('filter-domain').value;
  const source = document.getElementById('filter-source').value;
  const search = document.getElementById('filter-search').value;

  let query = 'events?limit=100';
  if (domain) query += `&domain=${domain}`;
  if (source) query += `&source_type=${source}`;

  state.events = await api.get(query);

  // Apply client-side search filter
  if (search) {
    const searchLower = search.toLowerCase();
    state.events = state.events.filter(e =>
      e.domain.toLowerCase().includes(searchLower) ||
      e.path.toLowerCase().includes(searchLower) ||
      e.event_type.toLowerCase().includes(searchLower)
    );
  }

  document.getElementById('events-table').innerHTML = renderEventsTable(state.events);
}

// Load Redirects management
async function loadRedirects() {
  const content = document.getElementById('app-content');
  state.redirects = await api.get('redirects');

  content.innerHTML = `
    <div class="row mb-3">
      <div class="col">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Create Redirect</h3>
          </div>
          <div class="card-body">
            <div class="row g-2">
              <div class="col-md-3">
                <input type="text" class="form-control" id="redirect-slug" placeholder="Slug (e.g., github)">
              </div>
              <div class="col-md-5">
                <input type="url" class="form-control" id="redirect-destination" placeholder="Destination URL">
              </div>
              <div class="col-md-2">
                <input type="text" class="form-control" id="redirect-tags" placeholder="Tags (comma-sep)">
              </div>
              <div class="col-md-2">
                <button class="btn btn-primary w-100" onclick="createRedirect()">Create</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-header">
        <h3 class="card-title">Redirects</h3>
      </div>
      <div class="card-table table-responsive">
        <table class="table table-vcenter">
          <thead>
            <tr>
              <th>Slug</th>
              <th>Destination</th>
              <th>Tags</th>
              <th>Clicks</th>
              <th>URL</th>
            </tr>
          </thead>
          <tbody id="redirects-table">
            ${renderRedirectsTable(state.redirects)}
          </tbody>
        </table>
      </div>
    </div>
  `;
}

function renderRedirectsTable(redirects) {
  return redirects.map(r => `
    <tr>
      <td><strong>${r.slug}</strong></td>
      <td class="table-truncate">${r.destination}</td>
      <td>${r.tags.filter(t => t).map(t => `<span class="tag-badge">${t}</span>`).join(' ')}</td>
      <td><span class="badge bg-blue">${r.click_count}</span></td>
      <td>
        <code class="table-truncate">/r/${r.slug}</code>
        <button class="btn btn-sm btn-ghost-secondary" onclick="copyToClipboard('${window.location.origin}/r/${r.slug}')">
          <i class="ti ti-copy"></i>
        </button>
      </td>
    </tr>
  `).join('');
}

async function createRedirect() {
  const slug = document.getElementById('redirect-slug').value;
  const destination = document.getElementById('redirect-destination').value;
  const tagsStr = document.getElementById('redirect-tags').value;
  const tags = tagsStr ? tagsStr.split(',').map(t => t.trim()) : [];

  if (!slug || !destination) {
    alert('Slug and destination are required');
    return;
  }

  try {
    await api.post('redirects', { slug, destination, tags });
    alert('Redirect created successfully!');
    await loadRedirects();
  } catch (error) {
    alert('Error: ' + error.message);
  }
}

// Load Webhooks configuration
async function loadWebhooks() {
  const content = document.getElementById('app-content');
  state.webhooks = await api.get('webhooks');

  content.innerHTML = `
    <div class="row mb-3">
      <div class="col">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Create Webhook</h3>
          </div>
          <div class="card-body">
            <div class="row g-2">
              <div class="col-md-4">
                <input type="text" class="form-control" id="webhook-name" placeholder="Name">
              </div>
              <div class="col-md-3">
                <input type="text" class="form-control" id="webhook-endpoint" placeholder="Endpoint">
              </div>
              <div class="col-md-3">
                <input type="text" class="form-control" id="webhook-secret" placeholder="Secret (optional)">
              </div>
              <div class="col-md-2">
                <button class="btn btn-primary w-100" onclick="createWebhook()">Create</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-header">
        <h3 class="card-title">Webhooks</h3>
      </div>
      <div class="card-table table-responsive">
        <table class="table table-vcenter">
          <thead>
            <tr>
              <th>Name</th>
              <th>Endpoint</th>
              <th>Secured</th>
              <th>Status</th>
              <th>URL</th>
            </tr>
          </thead>
          <tbody id="webhooks-table">
            ${renderWebhooksTable(state.webhooks)}
          </tbody>
        </table>
      </div>
    </div>
  `;
}

function renderWebhooksTable(webhooks) {
  return webhooks.map(w => `
    <tr>
      <td><strong>${w.name}</strong></td>
      <td><code>${w.endpoint}</code></td>
      <td>${w.has_secret ? '<i class="ti ti-lock text-success"></i>' : '<i class="ti ti-lock-open text-muted"></i>'}</td>
      <td>${w.is_active ? '<span class="badge bg-success">Active</span>' : '<span class="badge bg-secondary">Inactive</span>'}</td>
      <td>
        <code class="table-truncate">/webhook/${w.endpoint}</code>
        <button class="btn btn-sm btn-ghost-secondary" onclick="copyToClipboard('${window.location.origin}/webhook/${w.endpoint}')">
          <i class="ti ti-copy"></i>
        </button>
      </td>
    </tr>
  `).join('');
}

async function createWebhook() {
  const name = document.getElementById('webhook-name').value;
  const endpoint = document.getElementById('webhook-endpoint').value;
  const secret = document.getElementById('webhook-secret').value;

  if (!name || !endpoint) {
    alert('Name and endpoint are required');
    return;
  }

  try {
    await api.post('webhooks', { name, endpoint, secret });
    alert('Webhook created successfully!');
    await loadWebhooks();
  } catch (error) {
    alert('Error: ' + error.message);
  }
}

// Load Settings page
async function loadSettings() {
  const content = document.getElementById('app-content');

  content.innerHTML = `
    <div class="row row-cards">
      <div class="col-md-6">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Tracking Script</h3>
          </div>
          <div class="card-body">
            <p class="text-muted">Add this script to your website to start tracking:</p>
            <pre class="bg-light p-3"><code>&lt;script src="${window.location.origin}/static/js/track.min.js"&gt;&lt;/script&gt;</code></pre>
            <button class="btn btn-primary" onclick="copyToClipboard('&lt;script src=\\'${window.location.origin}/static/js/track.min.js\\'&gt;&lt;/script&gt;')">
              <i class="ti ti-copy"></i> Copy Script Tag
            </button>
          </div>
        </div>
      </div>

      <div class="col-md-6">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Tracking Pixel</h3>
          </div>
          <div class="card-body">
            <p class="text-muted">Add this pixel to track pageviews:</p>
            <pre class="bg-light p-3"><code>&lt;img src="${window.location.origin}/pixel.gif?domain=yoursite" style="display:none"&gt;</code></pre>
            <button class="btn btn-primary" onclick="copyToClipboard('&lt;img src=\\'${window.location.origin}/pixel.gif?domain=yoursite\\' style=\\'display:none\\'&gt;')">
              <i class="ti ti-copy"></i> Copy Pixel Code
            </button>
          </div>
        </div>
      </div>

      <div class="col-md-6">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">Appearance</h3>
          </div>
          <div class="card-body">
            <div class="form-group">
              <label class="form-label">Theme</label>
              <select class="form-select" id="theme-select" onchange="changeTheme(this.value)">
                <option value="light" ${state.theme === 'light' ? 'selected' : ''}>Light</option>
                <option value="dark" ${state.theme === 'dark' ? 'selected' : ''}>Dark</option>
              </select>
            </div>
          </div>
        </div>
      </div>

      <div class="col-md-6">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">About</h3>
          </div>
          <div class="card-body">
            <p><strong>Command Center</strong> v0.1.0</p>
            <p class="text-muted">Analytics & Monitoring Dashboard</p>
            <p class="text-muted">Port: 4698</p>
          </div>
        </div>
      </div>
    </div>
  `;
}

function changeTheme(theme) {
  state.theme = theme;
  document.body.setAttribute('data-bs-theme', theme);
  localStorage.setItem('cc-theme', theme);
  const icon = document.querySelector('#theme-toggle i');
  if (icon) icon.className = theme === 'light' ? 'ti ti-sun' : 'ti ti-moon';
}

function copyToClipboard(text) {
  const textarea = document.createElement('textarea');
  textarea.value = text;
  document.body.appendChild(textarea);
  textarea.select();
  document.execCommand('copy');
  document.body.removeChild(textarea);
  alert('Copied to clipboard!');
}

// Utility: Time ago formatter
function timeAgo(dateString) {
  const date = new Date(dateString);
  const seconds = Math.floor((new Date() - date) / 1000);

  if (seconds < 60) return 'Just now';
  if (seconds < 3600) return Math.floor(seconds / 60) + 'm ago';
  if (seconds < 86400) return Math.floor(seconds / 3600) + 'h ago';
  return Math.floor(seconds / 86400) + 'd ago';
}

// Theme toggle
function toggleTheme() {
  state.theme = state.theme === 'light' ? 'dark' : 'light';
  document.body.setAttribute('data-bs-theme', state.theme);
  localStorage.setItem('cc-theme', state.theme);

  const icon = document.querySelector('#theme-toggle i');
  icon.className = state.theme === 'light' ? 'ti ti-sun' : 'ti ti-moon';
}

// Refresh data
async function refreshData() {
  const btn = document.getElementById('refresh-btn');
  btn.classList.add('disabled');
  btn.innerHTML = '<div class="loading-spinner"></div>';

  try {
    await loadPage(state.currentPage);
  } finally {
    btn.classList.remove('disabled');
    btn.innerHTML = '<i class="ti ti-refresh"></i>';
  }
}

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
  // Set initial theme
  document.body.setAttribute('data-bs-theme', state.theme);
  const themeIcon = document.querySelector('#theme-toggle i');
  if (themeIcon) {
    themeIcon.className = state.theme === 'light' ? 'ti ti-sun' : 'ti ti-moon';
  }

  // Setup event listeners
  document.querySelectorAll('[data-page]').forEach(link => {
    link.addEventListener('click', (e) => {
      e.preventDefault();
      navigate(link.dataset.page);
    });
  });

  document.getElementById('theme-toggle')?.addEventListener('click', toggleTheme);
  document.getElementById('refresh-btn')?.addEventListener('click', refreshData);

  // Load initial page
  navigate('dashboard');

  // Register service worker for PWA
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/static/sw.js')
      .then(reg => console.log('Service Worker registered', reg))
      .catch(err => console.log('Service Worker registration failed', err));
  }

  // Auto-refresh every 30 seconds
  setInterval(() => {
    if (state.currentPage === 'dashboard') {
      refreshData();
    }
  }, 30000);
});
