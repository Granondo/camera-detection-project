<script>
  import { onMount, onDestroy } from 'svelte';
  import { api } from '../utils/api.js';
  import LoadingSpinner from './LoadingSpinner.svelte';
  
  let summary = null;
  let loading = true;
  let error = null;
  let refreshInterval;
  
  onMount(async () => {
    await loadAnalytics();
    refreshInterval = setInterval(loadAnalytics, 30000); // Каждые 30 секунд
  });
  
  onDestroy(() => {
    if (refreshInterval) {
      clearInterval(refreshInterval);
    }
  });
  
  async function loadAnalytics() {
    loading = true;
    error = null;
    
    try {
      const response = await api.request('/api/analytics/summary');
      
      if (response && response.success && response.data) {
        summary = response.data;
      } else {
        error = 'ClickHouse analytics not available';
      }
    } catch (err) {
      console.error('Error loading analytics:', err);
      error = err.message;
    } finally {
      loading = false;
    }
  }
</script>

<div class="analytics-dashboard">
  <div class="card">
    <div class="card-header">
      <h2 class="card-title">📊 ClickHouse Analytics</h2>
      <button class="btn-refresh" on:click={loadAnalytics}>
        🔄 Refresh
      </button>
    </div>
    
    {#if loading && !summary}
      <LoadingSpinner />
    {:else if error}
      <div class="error-box">
        ⚠️ {error}
      </div>
    {:else if summary}
      
      <!-- Stats -->
      <div class="analytics-stats">
        <div class="stat-box">
          <span class="stat-label">Total Detections</span>
          <span class="stat-value">{summary.total_detections.toLocaleString()}</span>
        </div>
        
        <div class="stat-box">
          <span class="stat-label">Detections Today</span>
          <span class="stat-value">{summary.detections_today.toLocaleString()}</span>
        </div>
        
        <div class="stat-box">
          <span class="stat-label">ClickHouse Status</span>
          <span class="stat-value" class:online={summary.clickhouse_available}>
            {summary.clickhouse_available ? '🟢 Online' : '🔴 Offline'}
          </span>
        </div>
      </div>
      
      <!-- Top Objects -->
      {#if summary.top_objects && summary.top_objects.length > 0}
        <div class="section">
          <h3>🎯 Top Detected Objects (Last 7 days)</h3>
          <div class="objects-list">
            {#each summary.top_objects as obj}
              <div class="object-item">
                <div class="object-info">
                  <span class="object-name">{obj.object_class}</span>
                  <span class="object-count">{obj.total_detections} detections</span>
                </div>
                <div class="object-confidence">
                  Avg: {(obj.avg_confidence * 100).toFixed(1)}%
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}
      
      <!-- Hourly Chart -->
      {#if summary.detections_by_hour && summary.detections_by_hour.length > 0}
        <div class="section">
          <h3>📈 Detections by Hour (Last 24h)</h3>
          <div class="chart-placeholder">
            <p>Chart visualization coming soon...</p>
            <small>Total: {summary.detections_by_hour.reduce((sum, h) => sum + h.total_detections, 0)} detections</small>
          </div>
        </div>
      {/if}
      
    {:else}
      <p class="empty-state">No analytics data available</p>
    {/if}
  </div>
</div>

<style>
  .analytics-dashboard {
    margin-block-end: var(--size-8);
  }
  
  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-block-end: var(--size-6);
  }
  
  .btn-refresh {
    padding: var(--size-2) var(--size-4);
    background: var(--primary);
    color: white;
    border: none;
    border-radius: var(--radius-2);
    cursor: pointer;
    font-size: var(--font-size-1);
  }
  
  .analytics-stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: var(--size-4);
    margin-block-end: var(--size-6);
  }
  
  .stat-box {
    background: linear-gradient(135deg, var(--purple-5), var(--indigo-6));
    color: white;
    padding: var(--size-4);
    border-radius: var(--radius-2);
    display: flex;
    flex-direction: column;
    gap: var(--size-2);
  }
  
  .stat-label {
    font-size: var(--font-size-1);
    opacity: 0.9;
  }
  
  .stat-value {
    font-size: var(--font-size-5);
    font-weight: var(--font-weight-7);
  }
  
  .stat-value.online {
    color: var(--green-4);
  }
  
  .section {
    margin-block-end: var(--size-6);
  }
  
  .section h3 {
    font-size: var(--font-size-3);
    color: var(--text-1);
    margin-block-end: var(--size-3);
  }
  
  .objects-list {
    display: flex;
    flex-direction: column;
    gap: var(--size-2);
  }
  
  .object-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--size-3);
    background: var(--surface-2);
    border-radius: var(--radius-2);
  }
  
  .object-info {
    display: flex;
    flex-direction: column;
    gap: var(--size-1);
  }
  
  .object-name {
    font-weight: var(--font-weight-6);
    text-transform: capitalize;
  }
  
  .object-count {
    font-size: var(--font-size-0);
    color: var(--text-2);
  }
  
  .object-confidence {
    font-size: var(--font-size-1);
    color: var(--text-2);
  }
  
  .chart-placeholder {
    background: var(--surface-2);
    padding: var(--size-8);
    border-radius: var(--radius-2);
    text-align: center;
    color: var(--text-2);
  }
  
  .error-box {
    background: var(--red-2);
    border: var(--border-size-1) solid var(--red-6);
    border-radius: var(--radius-2);
    padding: var(--size-4);
    color: var(--red-9);
  }
  
  .empty-state {
    text-align: center;
    color: var(--text-2);
    padding: var(--size-8);
  }
</style>