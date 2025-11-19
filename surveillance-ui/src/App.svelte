<script>
  import { onMount, onDestroy } from 'svelte';
  import { systemStore, isCameraOnline, cameraStatusText } from './lib/stores/systemStore.js';
  import { recordingsStore } from './lib/stores/recordingsStore.js';
  import { eventsStore } from './lib/stores/eventsStore.js';
  
  import StatCard from './lib/components/StatCard.svelte';
  import StatusIndicator from './lib/components/StatusIndicator.svelte';
  import EventCard from './lib/components/EventCard.svelte';
  import LoadingSpinner from './lib/components/LoadingSpinner.svelte';
  import AnalyticsDashboard from './lib/components/AnalyticsDashboard.svelte';
  
  import { formatBytes } from './lib/utils/formatters.js';
  import './app.css';

  let refreshInterval;
  
  onMount(async () => {
    await loadData();
    refreshInterval = setInterval(loadData, 10000);
  });
  
  onDestroy(() => {
    if (refreshInterval) {
      clearInterval(refreshInterval);
    }
  });
  
  async function loadData() {
    try {
      await systemStore.fetchStatus();
      await recordingsStore.fetch(10);
      await eventsStore.fetch(10);
    } catch (error) {
      console.error('Error loading data:', error);
    }
  }
  
  function handleRefresh() {
    loadData();
  }
</script>

<div class="app">
  <!-- Header -->
  <header class="header">
    <div class="container header-content">
      <div class="header-brand">
        <span class="header-icon">📹</span>
        <h1 class="header-title gradient-text">
          Camera Surveillance
        </h1>
      </div>
      
      <div class="header-actions">
        <StatusIndicator 
          online={$isCameraOnline} 
          label={$cameraStatusText} 
        />
        <button class="btn-refresh" on:click={handleRefresh}>
          🔄 Refresh
        </button>
      </div>
    </div>
  </header>

  <!-- Main Content -->
  <main class="main">
    <div class="container">
      
      {#if $systemStore.loading && !$systemStore.stats}
        <LoadingSpinner size="lg" />
      {:else if $systemStore.error}
        <div class="error-box">
          ❌ Error: {$systemStore.error}
        </div>
      {:else}
        
        <!-- Stats Cards -->
        <div class="stats-grid">
          <StatCard 
            title="Camera Status"
            value={$isCameraOnline ? 'Online' : 'Offline'}
            icon={$isCameraOnline ? '🟢' : '🔴'}
            color={$isCameraOnline ? 'green' : 'red'}
            loading={$systemStore.loading}
          />
          
          <StatCard 
            title="Total Recordings"
            value={$systemStore.stats?.recordings || 0}
            icon="📼"
            color="blue"
            loading={$systemStore.loading}
          />
          
          <StatCard 
            title="Frames Captured"
            value={$systemStore.stats?.frames || 0}
            icon="🖼️"
            color="orange"
            loading={$systemStore.loading}
          />
          
          <StatCard 
            title="Detections"
            value={$systemStore.stats?.detections || 0}
            icon="🎯"
            color="primary"
            loading={$systemStore.loading}
          />
        </div>

        <AnalyticsDashboard />

        <!-- Two Column Layout -->
        <div class="content-grid">
          
          <!-- Recent Events -->
          <div class="card">
            <div class="card-header">
              <h2 class="card-title">📢 Recent Events</h2>
              <span class="card-badge">{$eventsStore.events?.length || 0} events</span>
            </div>
            
            {#if $eventsStore.loading}
              <LoadingSpinner />
            {:else if $eventsStore.events && $eventsStore.events.length > 0}
              <div class="events-list">
                {#each $eventsStore.events as event, index (event.id || index)}
                  <EventCard {event} />
                {/each}
              </div>
            {:else}
              <p class="empty-state">No events yet</p>
            {/if}
          </div>

          <!-- System Info Column -->
          <div class="info-column">
            
            <!-- Camera Info -->
            {#if $systemStore.camera}
              <div class="card">
                <h2 class="card-title">📹 Camera Info</h2>
                <div class="info-list">
                  <div class="info-item">
                    <span class="info-label">Name:</span>
                    <span class="info-value">{$systemStore.camera.name}</span>
                  </div>
                  <div class="info-item">
                    <span class="info-label">Status:</span>
                    <StatusIndicator 
                      online={$isCameraOnline} 
                      label={$systemStore.camera.status} 
                    />
                  </div>
                  <div class="info-item">
                    <span class="info-label">Last Ping:</span>
                    <span class="info-value-small">
                      {$systemStore.camera.last_ping ? new Date($systemStore.camera.last_ping).toLocaleString() : 'Never'}
                    </span>
                  </div>
                </div>
              </div>
            {/if}

            <!-- Storage Info -->
            <div class="card">
              <h2 class="card-title">💾 Storage</h2>
              <div class="info-list">
                <div class="info-item">
                  <span class="info-label">Used Space:</span>
                  <span class="info-value">
                    {formatBytes($systemStore.stats?.storage_usage_bytes || 0)}
                  </span>
                </div>
                <div class="progress-bar">
                  <div 
                    class="progress-fill"
                    style="width: {Math.min((($systemStore.stats?.storage_usage_bytes || 0) / 10737418240) * 100, 100)}%"
                  ></div>
                </div>
              </div>
            </div>

            <!-- Recent Recordings -->
            <div class="card">
              <h2 class="card-title">📼 Recent Recordings</h2>
              
              {#if $recordingsStore.loading}
                <LoadingSpinner />
              {:else if $recordingsStore.recordings && $recordingsStore.recordings.length > 0}
                <div class="recordings-list">
                  {#each $recordingsStore.recordings.slice(0, 5) as rec, index (index)}
                    <div class="recording-item">
                      <div class="recording-info">
                        <p class="recording-title">Recording #{rec.recording?.id || 'N/A'}</p>
                        <p class="recording-time">
                          {rec.recording?.start_time ? new Date(rec.recording.start_time).toLocaleString() : 'Unknown'}
                        </p>
                      </div>
                      <span class="recording-size">{formatBytes(rec.recording?.file_size || 0)}</span>
                    </div>
                  {/each}
                </div>
              {:else}
                <p class="empty-state">No recordings yet</p>
              {/if}
            </div>

          </div>
        </div>
      {/if}
    </div>
  </main>

  <!-- Footer -->
  <footer class="footer">
    <div class="container">
      <p class="footer-text">
        Camera Surveillance System • Built with Svelte + Go + YOLO
      </p>
    </div>
  </footer>
</div>

<style>
  .app {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
  }
  
  /* Container */
  .container {
    width: 100%;
    max-width: var(--size-content-3);
    margin-inline: auto;
    padding-inline: var(--size-4);
  }
  
  /* Header */
  .header {
    background: var(--surface-1);
    box-shadow: var(--shadow-1);
  }
  
  .header-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-block: var(--size-4);
  }
  
  .header-brand {
    display: flex;
    align-items: center;
    gap: var(--size-3);
  }
  
  .header-icon {
    font-size: var(--font-size-6);
  }
  
  .header-title {
    font-size: var(--font-size-5);
    font-weight: var(--font-weight-7);
    margin: 0;
  }
  
  .header-actions {
    display: flex;
    align-items: center;
    gap: var(--size-4);
  }
  
  .btn-refresh {
    padding: var(--size-2) var(--size-4);
    background: var(--primary);
    color: white;
    border: none;
    border-radius: var(--radius-2);
    font-size: var(--font-size-1);
    font-weight: var(--font-weight-5);
    cursor: pointer;
    transition: background var(--animation-duration-fast) var(--ease-3);
    display: flex;
    align-items: center;
    gap: var(--size-2);
  }
  
  .btn-refresh:hover {
    background: var(--primary-dark);
  }
  
  /* Main */
  .main {
    flex: 1;
    padding-block: var(--size-8);
  }
  
  /* Stats Grid */
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: var(--size-6);
    margin-block-end: var(--size-8);
  }
  
  /* Content Grid */
  .content-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--size-8);
  }
  
  @media (min-width: 1024px) {
    .content-grid {
      grid-template-columns: 1fr 1fr;
    }
  }
  
  /* Info Column */
  .info-column {
    display: flex;
    flex-direction: column;
    gap: var(--size-6);
  }
  
  /* Card */
  .card {
    background: var(--surface-card);
    border-radius: var(--radius-3);
    box-shadow: var(--shadow-2);
    padding: var(--size-6);
  }
  
  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-block-end: var(--size-4);
  }
  
  .card-title {
    font-size: var(--font-size-4);
    font-weight: var(--font-weight-7);
    color: var(--text-1);
    margin: 0;
    margin-block-end: var(--size-4);
  }
  
  .card-badge {
    font-size: var(--font-size-1);
    color: var(--text-2);
  }
  
  /* Events List */
  .events-list {
    display: flex;
    flex-direction: column;
    gap: var(--size-3);
    max-height: 400px;
    overflow-y: auto;
  }
  
  /* Info List */
  .info-list {
    display: flex;
    flex-direction: column;
    gap: var(--size-3);
  }
  
  .info-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .info-label {
    color: var(--text-2);
    font-size: var(--font-size-1);
  }
  
  .info-value {
    font-weight: var(--font-weight-6);
    font-size: var(--font-size-1);
  }
  
  .info-value-small {
    font-size: var(--font-size-0);
    color: var(--text-2);
  }
  
  /* Progress Bar */
  .progress-bar {
    width: 100%;
    height: var(--size-2);
    background: var(--surface-3);
    border-radius: var(--radius-round);
    overflow: hidden;
    margin-block-start: var(--size-2);
  }
  
  .progress-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--purple-5), var(--blue-6));
    border-radius: var(--radius-round);
    transition: width var(--animation-duration-normal) var(--ease-3);
  }
  
  /* Recordings List */
  .recordings-list {
    display: flex;
    flex-direction: column;
    gap: var(--size-2);
  }
  
  .recording-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--size-3);
    background: var(--surface-2);
    border-radius: var(--radius-2);
    transition: background var(--animation-duration-fast) var(--ease-3);
  }
  
  .recording-item:hover {
    background: var(--surface-3);
  }
  
  .recording-info {
    flex: 1;
  }
  
  .recording-title {
    font-weight: var(--font-weight-5);
    font-size: var(--font-size-1);
    margin: 0;
    margin-block-end: var(--size-1);
  }
  
  .recording-time {
    font-size: var(--font-size-0);
    color: var(--text-2);
    margin: 0;
  }
  
  .recording-size {
    font-size: var(--font-size-1);
    color: var(--text-2);
  }
  
  /* Empty State */
  .empty-state {
    text-align: center;
    color: var(--text-2);
    padding-block: var(--size-8);
  }
  
  /* Error Box */
  .error-box {
    background: var(--red-2);
    border: var(--border-size-1) solid var(--red-6);
    border-radius: var(--radius-2);
    padding: var(--size-4);
    color: var(--red-9);
  }
  
  /* Footer */
  .footer {
    background: var(--surface-1);
    border-block-start: var(--border-size-1) solid var(--surface-3);
    margin-block-start: var(--size-12);
  }
  
  .footer-text {
    text-align: center;
    color: var(--text-2);
    font-size: var(--font-size-1);
    padding-block: var(--size-4);
    margin: 0;
  }
</style>