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
        content.innerHTML = '<div class="alert alert-info">Analytics page coming soon...</div>';
        break;
      case 'redirects':
        content.innerHTML = '<div class="alert alert-info">Redirects management coming soon...</div>';
        break;
      case 'webhooks':
        content.innerHTML = '<div class="alert alert-info">Webhooks configuration coming soon...</div>';
        break;
      case 'settings':
        content.innerHTML = '<div class="alert alert-info">Settings page coming soon...</div>';
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

  // Auto-refresh every 30 seconds
  setInterval(() => {
    if (state.currentPage === 'dashboard') {
      refreshData();
    }
  }, 30000);
});
