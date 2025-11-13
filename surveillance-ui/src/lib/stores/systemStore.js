import { writable, derived } from 'svelte/store';
import { api } from '../utils/api.js';

function createSystemStore() {
  const { subscribe, set, update } = writable({
    status: null,
    stats: null,
    camera: null,
    loading: true,
    error: null,
  });

  return {
    subscribe,
    
    fetchStatus: async function() {
      console.log('📊 Fetching system status...');
      update(state => ({ ...state, loading: true, error: null }));
      
      try {
        const response = await api.getStatus();
        console.log('✅ Status response:', response);
        
        if (response && response.success && response.data) {
          update(state => ({
            ...state,
            status: response.data.system_status,
            stats: response.data.stats,
            camera: response.data.camera,
            loading: false,
          }));
          console.log('✅ System status updated');
        } else if (response && response.cached) {
          console.log('ℹ️ Using cached data');
          update(state => ({ ...state, loading: false }));
        } else {
          throw new Error('Invalid response from API');
        }
      } catch (error) {
        console.error('❌ Error fetching status:', error);
        update(state => ({
          ...state,
          error: error.message,
          loading: false,
        }));
      }
    },

    reset: function() {
      set({
        status: null,
        stats: null,
        camera: null,
        loading: true,
        error: null,
      });
    }
  };
}

export const systemStore = createSystemStore();

// ✅ ИСПРАВЛЕННАЯ ЛОГИКА
export const isCameraOnline = derived(
  systemStore,
  $system => {
    if (!$system.camera) return false;
    
    // Проверяем: active + last_ping был недавно (< 5 минут)
    const isActive = $system.camera.status === 'active';
    
    if (!isActive) return false;
    
    if (!$system.camera.last_ping) return false;
    
    const lastPingTime = new Date($system.camera.last_ping).getTime();
    const now = Date.now();
    const fiveMinutes = 5 * 60 * 1000;
    
    return (now - lastPingTime) < fiveMinutes;
  }
);

// ✅ Дополнительный derived store для текста статуса
export const cameraStatusText = derived(
  [systemStore, isCameraOnline],
  ([$system, $online]) => {
    if (!$system.camera) return 'Unknown';
    
    if ($online) return 'Online';
    
    if ($system.camera.status === 'active' && $system.camera.last_ping) {
      const lastPingTime = new Date($system.camera.last_ping).getTime();
      const now = Date.now();
      const diffMs = now - lastPingTime;
      const diffMinutes = Math.floor(diffMs / 1000 / 60);
      
      if (diffMinutes < 60) {
        return `Offline (${diffMinutes}m ago)`;
      } else if (diffMinutes < 1440) {
        return `Offline (${Math.floor(diffMinutes / 60)}h ago)`;
      } else {
        return `Offline (${Math.floor(diffMinutes / 1440)}d ago)`;
      }
    }
    
    return 'Offline';
  }
);