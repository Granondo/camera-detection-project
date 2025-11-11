import { writable, derived } from 'svelte/store';
import { api } from '../utils/api.js';

// System status store
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

// Derived store for camera online status
export const isCameraOnline = derived(
  systemStore,
  $system => $system.status === 'online'
);