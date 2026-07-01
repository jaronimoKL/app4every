import re

with open('frontend/src/views/watchparty/WatchPartyRoom.vue', 'r') as f:
    content = f.read()

# Remove the duplicates at the top of script setup (around line 351-358)
content = re.sub(
    r"import \{ useAuthStore \} from '@\/stores\/auth'\nconst authStore = useAuthStore\(\)\n",
    r"",
    content
)

content = re.sub(
    r"import \{ useReviewsStore \} from '@\/stores\/reviews'\nconst reviewsStore = useReviewsStore\(\)\n",
    r"",
    content
)

# Ensure useAuthStore is imported alongside useWatchParty
if "import { useAuthStore }" not in content:
    content = content.replace(
        "import { useWatchParty } from '@/composables/useWatchParty'",
        "import { useWatchParty } from '@/composables/useWatchParty'\nimport { useAuthStore } from '@/stores/auth'"
    )

# Fix syncProgressWithList to use `auth` instead of `authStore`
content = content.replace('authStore.isAuthenticated', 'auth.isAuthenticated')

with open('frontend/src/views/watchparty/WatchPartyRoom.vue', 'w') as f:
    f.write(content)

