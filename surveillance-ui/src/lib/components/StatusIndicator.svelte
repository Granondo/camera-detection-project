<script>
  export let online = false;
  export let label = 'Status';
</script>

<div class="status-indicator">
  <div class="status-dot" class:online>
    {#if online}
      <div class="pulse-ring"></div>
    {/if}
  </div>
  <span class="status-label" class:online>{label}</span>
</div>

<style>
  .status-indicator {
    display: flex;
    align-items: center;
    gap: var(--size-2);
  }
  
  .status-dot {
    position: relative;
    width: var(--size-3);
    height: var(--size-3);
    border-radius: var(--radius-round);
    background: var(--red-6);
  }
  
  .status-dot.online {
    background: var(--green-6);
  }
  
  .pulse-ring {
    position: absolute;
    inset: 0;
    width: var(--size-3);
    height: var(--size-3);
    border-radius: var(--radius-round);
    background: var(--green-6);
    animation: pulse 2s var(--ease-3) infinite;
  }
  
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }
  
  .status-label {
    font-size: var(--font-size-1);
    font-weight: var(--font-weight-5);
    color: var(--red-7);
  }
  
  .status-label.online {
    color: var(--green-7);
  }
</style>