<script>
  import { formatDateTime, formatBytes, formatDuration } from '../utils/formatters';
  import api from '../utils/api';
  
  export let recording;
  
  const rec = recording.recording;
  const videoUrl = api.getVideoUrl(rec.id);
  const downloadUrl = api.getDownloadUrl(rec.id);
</script>

<article class="recording">
  <div class="info">
    <h4>Recording #{rec.id}</h4>
    <div class="meta">
      <span>📅 {formatDateTime(rec.start_time)}</span>
      <span>⏱️ {formatDuration(rec.duration)}</span>
      <span>💾 {formatBytes(rec.file_size)}</span>
    </div>
  </div>
  
  <div class="actions">
    {#if recording.file_exists}
      <a href={videoUrl} class="btn btn-sm" target="_blank">
        ▶️ Play
      </a>
      <a href={downloadUrl} class="btn btn-sm" download>
        ⬇️ Download
      </a>
    {:else}
      <span class="text-muted text-small">File missing</span>
    {/if}
  </div>
</article>

<style>
  .recording {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--size-4);
    padding: var(--size-3);
    background: var(--surface-2);
    border-radius: var(--radius-2);
    transition: background 0.2s var(--ease-3);
  }
  
  .recording:hover {
    background: var(--surface-3);
  }
  
  .info {
    flex: 1;
    min-width: 0;
  }
  
  h4 {
    font-size: var(--font-size-2);
    font-weight: var(--font-weight-6);
    margin: 0 0 var(--size-1) 0;
    color: var(--text-1);
  }
  
  .meta {
    display: flex;
    flex-wrap: wrap;
    gap: var(--size-3);
    font-size: var(--font-size-0);
    color: var(--text-3);
  }
  
  .actions {
    display: flex;
    gap: var(--size-2);
  }
  
  .btn-sm {
    padding: var(--size-1) var(--size-3);
    font-size: var(--font-size-1);
  }
</style>