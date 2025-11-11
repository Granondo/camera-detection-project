<script>
  export let title = '';
  export let value = '';
  export let icon = '📊';
  export let color = 'primary';
  export let loading = false;
</script>

<div class="stat-card" data-color={color}>
  <div class="stat-card-content">
    <div class="stat-card-info">
      <p class="stat-card-title">{title}</p>
      {#if loading}
        <div class="loading-skeleton"></div>
      {:else}
        <p class="stat-card-value">{value}</p>
      {/if}
    </div>
    <div class="stat-card-icon" data-color={color}>
      <span>{icon}</span>
    </div>
  </div>
</div>

<style>
  .stat-card {
    background: var(--surface-card);
    border-radius: var(--radius-3);
    box-shadow: var(--shadow-2);
    padding: var(--size-6);
    transition: box-shadow var(--animation-duration-normal) var(--ease-3);
  }
  
  .stat-card:hover {
    box-shadow: var(--shadow-4);
  }
  
  .stat-card-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--size-4);
  }
  
  .stat-card-info {
    flex: 1;
  }
  
  .stat-card-title {
    color: var(--text-2);
    font-size: var(--font-size-1);
    font-weight: var(--font-weight-5);
    margin-block-end: var(--size-2);
  }
  
  .stat-card-value {
    font-size: var(--font-size-6);
    font-weight: var(--font-weight-7);
    color: var(--text-1);
    margin: 0;
  }
  
  .stat-card-icon {
    padding: var(--size-4);
    border-radius: var(--radius-3);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .stat-card-icon[data-color="primary"] {
    background: linear-gradient(135deg, var(--purple-5), var(--indigo-6));
  }
  
  .stat-card-icon[data-color="green"] {
    background: linear-gradient(135deg, var(--green-5), var(--emerald-6));
  }
  
  .stat-card-icon[data-color="blue"] {
    background: linear-gradient(135deg, var(--blue-5), var(--cyan-6));
  }
  
  .stat-card-icon[data-color="orange"] {
    background: linear-gradient(135deg, var(--orange-5), var(--amber-6));
  }
  
  .stat-card-icon[data-color="red"] {
    background: linear-gradient(135deg, var(--red-5), var(--rose-6));
  }
  
  .stat-card-icon span {
    font-size: var(--font-size-6);
  }
  
  .loading-skeleton {
    width: var(--size-10);
    height: var(--size-8);
    background: var(--surface-3);
    border-radius: var(--radius-2);
    animation: pulse 2s var(--ease-3) infinite;
  }
</style>