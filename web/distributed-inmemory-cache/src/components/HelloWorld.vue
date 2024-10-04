<template>
  <div class="container">
    <h1>Data Management Dashboard</h1>
    <div class="tabs">
      <button
          v-for="tab in tabs"
          :key="tab"
          :class="['tab', { active: currentTab === tab }]"
          @click="currentTab = tab"
      >
        {{ tab }}
      </button>
    </div>

    <div class="tab-content">
      <!-- View Data Tab -->
      <div v-if="currentTab === 'View Data'">
        <div class="header">
          <h2>View Data</h2>
          <button @click="fetchData" :disabled="loadingData" class="refresh-button">
            Refresh
          </button>
        </div>
        <div v-if="loadingData">Loading...</div>
        <div v-else>
          <pre>{{ formattedData }}</pre>
        </div>
      </div>

      <!-- Set Data Tab -->
      <div v-if="currentTab === 'Set Data'">
        <div class="header">
          <h2>Set Data</h2>
          <button @click="addNewPair" class="add-button">Add New</button>
        </div>
        <form @submit.prevent="setData">
          <div
              class="key-value-pair"
              v-for="(pair, index) in keyValuePairs"
              :key="index"
          >
            <div class="form-group">
              <label :for="`key-${index}`">Key:</label>
              <input
                  v-model="pair.key"
                  :id="`key-${index}`"
                  required
                  placeholder="Enter key"
              />
            </div>
            <div class="form-group">
              <label :for="`value-${index}`">Value:</label>
              <input
                  v-model="pair.value"
                  :id="`value-${index}`"
                  required
                  placeholder="Enter value"
              />
            </div>
            <button
                type="button"
                class="remove-button"
                @click="removePair(index)"
                v-if="keyValuePairs.length > 1"
            >
              Remove
            </button>
          </div>
          <button type="submit" class="submit-button">Set Data</button>
        </form>
        <div v-if="setDataMessage" :class="{'message': true, 'error': setDataError}">
          {{ setDataMessage }}
        </div>
      </div>

      <!-- Delete Data Tab -->
      <div v-if="currentTab === 'Delete Data'">
        <h2>Delete Data</h2>
        <form @submit.prevent="deleteData">
          <div class="form-group">
            <label for="deleteKey">Key:</label>
            <input v-model="deleteKey" id="deleteKey" required placeholder="Enter key to delete" />
          </div>
          <button type="submit" class="submit-button">Delete Data</button>
        </form>
        <div v-if="deleteDataMessage" :class="{'message': true, 'error': deleteDataError}">
          {{ deleteDataMessage }}
        </div>
      </div>

      <!-- Node Statistics Tab -->
      <div v-if="currentTab === 'Node Statistics'">
        <div class="header">
          <h2>Node Statistics</h2>
          <button @click="fetchNodeStats" :disabled="loadingStats" class="refresh-button">
            Refresh
          </button>
        </div>
        <div v-if="loadingStats">Loading...</div>
        <div v-else>
          <pre>{{ formattedNodeStats }}</pre>
        </div>
      </div>

      <!-- Infra Management Tab -->
      <div v-if="currentTab === 'Infra Management'">
        <h2>Infrastructure Management</h2>
        <div class="button-group">
          <button @click="scaleUp" :disabled="infraLoading">Scale Up</button>
          <button @click="scaleDown" :disabled="infraLoading">Scale Down</button>
          <button @click="confirmKillAll" :disabled="infraLoading">Kill All</button>
        </div>
        <div v-if="infraMessage" :class="{'message': true, 'error': infraError}">
          {{ infraMessage }}
        </div>
        <div v-if="infraLoading" class="loading">Processing...</div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, computed } from 'vue';
import axios from 'axios';

export default {
  name: 'DataManagement',
  setup() {
    const tabs = ['View Data', 'Set Data', 'Delete Data', 'Node Statistics', 'Infra Management'];
    const currentTab = ref('View Data');

    // -------------------
    // View Data Section
    // -------------------
    const data = ref(null);
    const loadingData = ref(false);

    const fetchData = async () => {
      loadingData.value = true;
      try {
        const response = await axios.get('/api/data/get');
        data.value = response.data;
      } catch (error) {
        data.value = 'Error fetching data.';
        console.error(error);
      } finally {
        loadingData.value = false;
      }
    };

    const formattedData = computed(() => {
      return typeof data.value === 'object' ? JSON.stringify(data.value, null, 2) : data.value;
    });

    // -------------------
    // Set Data Section
    // -------------------
    const keyValuePairs = ref([{ key: '', value: '' }]);
    const setDataMessage = ref('');
    const setDataError = ref(false);

    const addNewPair = () => {
      keyValuePairs.value.push({ key: '', value: '' });
    };

    const removePair = (index) => {
      keyValuePairs.value.splice(index, 1);
    };

    const setData = async () => {
      setDataMessage.value = '';
      setDataError.value = false;

      // Construct the payload
      const payload = {};
      for (const pair of keyValuePairs.value) {
        if (pair.key.trim() === '' || pair.value.trim() === '') {
          setDataMessage.value = 'All key-value pairs must be filled.';
          setDataError.value = true;
          return;
        }
        payload[pair.key] = pair.value;
      }

      try {
        await axios.post('/api/data/set', payload);
        setDataMessage.value = 'Data set successfully!';
        keyValuePairs.value = [{ key: '', value: '' }]; // Reset form
        // Optionally refresh the view data
        if (currentTab.value === 'View Data') {
          fetchData();
        }
      } catch (error) {
        setDataMessage.value = 'Error setting data.';
        setDataError.value = true;
        console.error(error);
      }
    };

    // -------------------
    // Delete Data Section
    // -------------------
    const deleteKey = ref('');
    const deleteDataMessage = ref('');
    const deleteDataError = ref(false);

    const deleteData = async () => {
      deleteDataMessage.value = '';
      deleteDataError.value = false;
      try {
        await axios.post('/api/data/delete', [deleteKey.value]);
        deleteDataMessage.value = 'Data deleted successfully!';
        deleteKey.value = ''; // Reset form
        // Optionally refresh the view data
        if (currentTab.value === 'View Data') {
          fetchData();
        }
      } catch (error) {
        deleteDataMessage.value = 'Error deleting data.';
        deleteDataError.value = true;
        console.error(error);
      }
    };

    // -------------------
    // Node Statistics Section
    // -------------------
    const nodeStats = ref(null);
    const loadingStats = ref(false);

    const fetchNodeStats = async () => {
      loadingStats.value = true;
      try {
        const response = await axios.get('/api/infra/nodestats');
        nodeStats.value = response.data;
      } catch (error) {
        nodeStats.value = 'Error fetching node statistics.';
        console.error(error);
      } finally {
        loadingStats.value = false;
      }
    };

    const formattedNodeStats = computed(() => {
      return typeof nodeStats.value === 'object' ? JSON.stringify(nodeStats.value, null, 2) : nodeStats.value;
    });

    // -------------------
    // Infra Management Section
    // -------------------
    const infraMessage = ref('');
    const infraError = ref(false);
    const infraLoading = ref(false);

    const scaleUp = async () => {
      await performInfraAction('/api/infra/scaleup', 'Scale Up');
    };

    const scaleDown = async () => {
      await performInfraAction('/api/infra/scaledown', 'Scale Down');
    };

    const killAll = async () => {
      await performInfraAction('/api/infra/killall', 'Kill All');
    };

    const confirmKillAll = () => {
      if (confirm('Are you sure you want to kill all infra processes? This action cannot be undone.')) {
        killAll();
      }
    };

    const performInfraAction = async (url, actionName) => {
      infraMessage.value = '';
      infraError.value = false;
      infraLoading.value = true;
      try {
        await axios.post(url);
        infraMessage.value = `${actionName} action completed successfully!`;
        // Optionally refresh node stats
        if (currentTab.value === 'Node Statistics') {
          fetchNodeStats();
        }
      } catch (error) {
        infraMessage.value = `Error performing ${actionName} action.`;
        infraError.value = true;
        console.error(error);
      } finally {
        infraLoading.value = false;
      }
    };

    // -------------------
    // Lifecycle Hook
    // -------------------
    onMounted(() => {
      fetchData();
      fetchNodeStats();
    });

    return {
      tabs,
      currentTab,
      // View Data
      data,
      loadingData,
      fetchData,
      formattedData,
      // Set Data
      keyValuePairs,
      addNewPair,
      removePair,
      setData,
      setDataMessage,
      setDataError,
      // Delete Data
      deleteKey,
      deleteData,
      deleteDataMessage,
      deleteDataError,
      // Node Statistics
      nodeStats,
      loadingStats,
      fetchNodeStats,
      formattedNodeStats,
      // Infra Management
      infraMessage,
      infraError,
      infraLoading,
      scaleUp,
      scaleDown,
      killAll,
      confirmKillAll,
    };
  },
};
</script>

<style scoped>
/* Dark Theme Styles */
.container {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
  background-color: #1e1e1e;
  color: #ffffff;
  min-height: 100vh;
  box-sizing: border-box;
}

h1 {
  text-align: center;
  margin-bottom: 20px;
  color: #ffffff;
}

.tabs {
  display: flex;
  border-bottom: 2px solid #444;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.tab {
  padding: 10px 20px;
  cursor: pointer;
  border: none;
  background: none;
  outline: none;
  transition: background-color 0.3s, color 0.3s;
  margin-right: 5px;
  margin-bottom: 5px;
  border-radius: 4px;
  color: #ffffff; /* White text */
}

.tab:hover {
  background-color: #333;
}

.tab.active {
  border-bottom: 2px solid #00aaff;
  font-weight: bold;
  background-color: #333;
}

.tab-content {
  padding: 20px;
  border: 1px solid #444;
  border-top: none;
  border-radius: 0 0 4px 4px;
  background-color: #2a2a2a;
}

/* Header for Tabs with Refresh/Add Buttons */
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.refresh-button,
.add-button {
  padding: 8px 12px;
  background-color: #00aaff;
  color: #ffffff;
  border: none;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.refresh-button:hover,
.add-button:hover {
  background-color: #0088cc;
}

.refresh-button:disabled,
.add-button:disabled {
  background-color: #555555;
  cursor: not-allowed;
}

/* Set Data Form Styles */
.key-value-pair {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
  flex-wrap: wrap;
}

.form-group {
  flex: 1;
  min-width: 150px;
  margin-right: 10px;
  margin-bottom: 10px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  color: #dddddd;
}

.form-group input {
  width: 100%;
  padding: 8px;
  box-sizing: border-box;
  background-color: #3a3a3a;
  border: 1px solid #555555;
  border-radius: 4px;
  color: #ffffff;
}

.remove-button {
  padding: 8px 12px;
  background-color: #ff5555;
  color: #ffffff;
  border: none;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.3s;
  height: fit-content;
}

.remove-button:hover {
  background-color: #cc4444;
}

.submit-button {
  padding: 10px 15px;
  background-color: #00aaff;
  color: #ffffff;
  border: none;
  cursor: pointer;
  border-radius: 4px;
  margin-top: 10px;
  transition: background-color 0.3s;
}

.submit-button:hover {
  background-color: #0088cc;
}

/* Delete Data Styles */
.delete-form {
  margin-bottom: 15px;
}

.delete-form .form-group input {
  width: 100%;
  padding: 8px;
  box-sizing: border-box;
  background-color: #3a3a3a;
  border: 1px solid #555555;
  border-radius: 4px;
  color: #ffffff;
}

/* Infra Management Styles */
.button-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.button-group button {
  flex: 1;
  min-width: 100px;
  padding: 10px 15px;
  background-color: #00aaff;
  color: #ffffff;
  border: none;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.button-group button:hover {
  background-color: #0088cc;
}

.button-group button:disabled {
  background-color: #555555;
  cursor: not-allowed;
}

/* Messages */
.message {
  margin-top: 15px;
  color: #00ff00; /* Green for success */
}

.message.error {
  color: #ff5555; /* Red for errors */
}

/* Loading Indicator */
.loading {
  margin-top: 15px;
  color: #00aaff;
}
</style>
