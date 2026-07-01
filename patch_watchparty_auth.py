with open('frontend/src/views/watchparty/WatchPartyRoom.vue', 'r') as f:
    content = f.read()

# Add imports and store initialization back at the top of script setup
content = content.replace('import { useRoute } from \'vue-router\'', 
'''import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
const authStore = useAuthStore()
''')

with open('frontend/src/views/watchparty/WatchPartyRoom.vue', 'w') as f:
    f.write(content)

