#!/usr/bin/env node

/**
 * Setup Test Worlds Script
 * 
 * Creates the test worlds needed for e2e testing.
 * Run this once before running e2e tests:
 * 
 * npm run setup-test-worlds
 * 
 * This will create persistent test worlds that can be reused across test runs.
 */

const SERVER_URL = 'http://localhost:8080';

const TEST_WORLDS = {
  'basic-movement': {
    name: 'Basic Movement Test World',
    description: '3x3 map with units for testing basic movement mechanics',
    creatorId: 'test-user',
    tiles: [
      {q: 0, r: 0, player: 0, tileType: 1}, {q: 1, r: 0, player: 0, tileType: 1}, {q: 2, r: 0, player: 0, tileType: 1},
      {q: 0, r: 1, player: 0, tileType: 1}, {q: 1, r: 1, player: 0, tileType: 1}, {q: 2, r: 1, player: 0, tileType: 1},
      {q: 0, r: 2, player: 0, tileType: 1}, {q: 1, r: 2, player: 0, tileType: 1}, {q: 2, r: 2, player: 0, tileType: 1}
    ],
    units: [
      {q: 0, r: 0, player: 1, unitType: 1, availableHealth: 100, distanceLeft: 3, turnCounter: 1},
      {q: 2, r: 2, player: 2, unitType: 1, availableHealth: 100, distanceLeft: 3, turnCounter: 1}
    ]
  },

  'combat-basic': {
    name: 'Basic Combat Test World',
    description: 'Adjacent units ready for combat testing',
    creatorId: 'test-user',
    tiles: [
      {q: 0, r: 0, player: 0, tileType: 1}, {q: 1, r: 0, player: 0, tileType: 1}, {q: 2, r: 0, player: 0, tileType: 1},
      {q: 0, r: 1, player: 0, tileType: 1}, {q: 1, r: 1, player: 0, tileType: 1}, {q: 2, r: 1, player: 0, tileType: 1}
    ],
    units: [
      {q: 0, r: 0, player: 1, unitType: 1, availableHealth: 100, distanceLeft: 3, turnCounter: 1},
      {q: 1, r: 0, player: 2, unitType: 1, availableHealth: 100, distanceLeft: 3, turnCounter: 1},
      {q: 0, r: 1, player: 1, unitType: 1, availableHealth: 100, distanceLeft: 3, turnCounter: 1}
    ]
  },

  'turn-flow': {
    name: 'Turn Management Test World',
    description: 'Multi-unit scenario for testing turn mechanics',
    creatorId: 'test-user',
    tiles: Array.from({length: 25}, (_, i) => ({
      q: i % 5,
      r: Math.floor(i / 5),
      player: 0,
      tileType: 1
    })),
    units: [
      {q: 0, r: 0, player: 1, unitType: 1, availableHealth: 100, distanceLeft: 3, turnCounter: 1},
      {q: 1, r: 0, player: 1, unitType: 1, availableHealth: 100, distanceLeft: 3, turnCounter: 1},
      {q: 3, r: 4, player: 2, unitType: 1, availableHealth: 100, distanceLeft: 3, turnCounter: 1},
      {q: 4, r: 4, player: 2, unitType: 1, availableHealth: 100, distanceLeft: 3, turnCounter: 1}
    ]
  },

  'error-handling': {
    name: 'Error Handling Test World',
    description: 'Constrained scenario for testing invalid moves and error conditions',
    creatorId: 'test-user',
    tiles: [
      {q: 0, r: 0, player: 0, tileType: 1}, {q: 1, r: 0, player: 0, tileType: 2}, // Mountain blocks path
      {q: 0, r: 1, player: 0, tileType: 1}, {q: 1, r: 1, player: 0, tileType: 1}
    ],
    units: [
      {q: 0, r: 0, player: 1, unitType: 1, availableHealth: 100, distanceLeft: 1, turnCounter: 1}, // Limited movement
      {q: 1, r: 1, player: 2, unitType: 1, availableHealth: 100, distanceLeft: 3, turnCounter: 1}
    ]
  }
};

async function checkServerHealth() {
  try {
    const response = await fetch(`${SERVER_URL}/`);
    if (!response.ok) {
      throw new Error(`Server health check failed: ${response.status}`);
    }
    console.log('✅ Server is running at', SERVER_URL);
    return true;
  } catch (error) {
    console.error(`❌ Server not accessible at ${SERVER_URL}`);
    console.error(`   Make sure to start the server: ./weewar-server --port=8080`);
    console.error(`   Error: ${error.message}`);
    return false;
  }
}

async function createTestWorld(worldId, worldData) {
  try {
    console.log(`🔧 Creating test world: ${worldId}`);
    
    const payload = {
      world: {
        id: worldId,
        name: worldData.name,
        description: worldData.description,
        creatorId: worldData.creatorId,
      },
      worldData: {
        tiles: worldData.tiles,
        units: worldData.units
      }
    };

    console.log(`   📄 Payload:`, JSON.stringify(payload, null, 2));

    const response = await fetch(`${SERVER_URL}/api/v1/worlds`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload)
    });

    if (!response.ok) {
      const errorText = await response.text();
      if (response.status === 409) {
        console.log(`   ⚠️  World ${worldId} already exists, skipping...`);
        return { success: true, worldId, skipped: true };
      } else {
        throw new Error(`Failed to create world: ${response.status} ${errorText}`);
      }
    }

    const result = await response.json();
    console.log(`   ✅ World created successfully: ${worldId}`);
    console.log(`   📊 Response:`, JSON.stringify(result, null, 2));
    
    return { success: true, worldId, data: result };

  } catch (error) {
    console.error(`   ❌ Failed to create world ${worldId}:`, error.message);
    return { success: false, worldId, error: error.message };
  }
}

async function deleteTestWorld(worldId) {
  try {
    console.log(`🗑️  Deleting test world: ${worldId}`);
    
    const response = await fetch(`${SERVER_URL}/api/v1/worlds/${worldId}`, {
      method: 'DELETE'
    });

    if (!response.ok) {
      if (response.status === 404) {
        console.log(`   ⚠️  World ${worldId} not found, skipping...`);
        return { success: true, worldId, skipped: true };
      } else {
        const errorText = await response.text();
        throw new Error(`Failed to delete world: ${response.status} ${errorText}`);
      }
    }

    console.log(`   ✅ World deleted successfully: ${worldId}`);
    return { success: true, worldId };

  } catch (error) {
    console.error(`   ❌ Failed to delete world ${worldId}:`, error.message);
    return { success: false, worldId, error: error.message };
  }
}

async function cleanupAllTestWorlds() {
  console.log('🧹 Cleaning up test worlds...\n');

  // Check server health first
  const serverOk = await checkServerHealth();
  if (!serverOk) {
    process.exit(1);
  }

  console.log('');

  // Delete each test world
  const results = [];
  for (const worldId of Object.keys(TEST_WORLDS)) {
    const result = await deleteTestWorld(worldId);
    results.push(result);
    console.log(''); // Add spacing between worlds
  }

  // Summary
  console.log('📋 Cleanup Summary:');
  const successful = results.filter(r => r.success);
  const failed = results.filter(r => !r.success);
  const skipped = results.filter(r => r.success && r.skipped);

  console.log(`   ✅ Successfully deleted: ${successful.length - skipped.length} worlds`);
  if (skipped.length > 0) {
    console.log(`   ⚠️  Not found: ${skipped.length} worlds`);
  }
  if (failed.length > 0) {
    console.log(`   ❌ Failed: ${failed.length} worlds`);
    failed.forEach(r => console.log(`      - ${r.worldId}: ${r.error}`));
  }

  // Remove the config file
  const fs = require('fs');
  const path = require('path');
  const configPath = path.join(__dirname, 'test-world-ids.json');
  try {
    fs.unlinkSync(configPath);
    console.log(`   📝 Removed config file: ${configPath}`);
  } catch (error) {
    console.log(`   ⚠️  Config file not found: ${configPath}`);
  }

  console.log('\n✨ Cleanup complete!');

  if (failed.length > 0) {
    process.exit(1);
  }
}

async function createAllTestWorlds() {
  console.log('🚀 Setting up test worlds for e2e testing...\n');

  // Check server health first
  const serverOk = await checkServerHealth();
  if (!serverOk) {
    process.exit(1);
  }

  console.log('');

  // Create each test world
  const results = [];
  for (const [worldId, worldData] of Object.entries(TEST_WORLDS)) {
    const result = await createTestWorld(worldId, worldData);
    results.push(result);
    console.log(''); // Add spacing between worlds
  }

  // Summary
  console.log('📋 Setup Summary:');
  const successful = results.filter(r => r.success);
  const failed = results.filter(r => !r.success);
  const skipped = results.filter(r => r.success && r.skipped);

  console.log(`   ✅ Successfully created: ${successful.length - skipped.length} worlds`);
  if (skipped.length > 0) {
    console.log(`   ⚠️  Already existed: ${skipped.length} worlds`);
  }
  if (failed.length > 0) {
    console.log(`   ❌ Failed: ${failed.length} worlds`);
    failed.forEach(r => console.log(`      - ${r.worldId}: ${r.error}`));
  }

  console.log('\n🎯 Test worlds ready! You can now run:');
  console.log('   npm run test:e2e');

  // Write world IDs to a file for tests to use
  const worldIds = {};
  successful.forEach(r => {
    worldIds[r.worldId] = r.worldId;
  });

  const fs = require('fs');
  const path = require('path');
  const configPath = path.join(__dirname, 'test-world-ids.json');
  fs.writeFileSync(configPath, JSON.stringify(worldIds, null, 2));
  console.log(`   📝 World IDs saved to: ${configPath}`);

  if (failed.length > 0) {
    process.exit(1);
  }
}

// Handle command line arguments
const args = process.argv.slice(2);
const isCleanup = args.includes('--cleanup');

// Run setup or cleanup
if (isCleanup) {
  cleanupAllTestWorlds().catch(error => {
    console.error('💥 Cleanup failed:', error);
    process.exit(1);
  });
} else {
  createAllTestWorlds().catch(error => {
    console.error('💥 Setup failed:', error);
    process.exit(1);
  });
}
