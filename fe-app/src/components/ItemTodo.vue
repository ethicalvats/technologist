<template>
    <li class="list-item">{{ item.text }}</li>
</template>

<script lang="ts">
import Vue from 'vue'
export default Vue.extend({
    props:{
        item: {
            type: Object,
            required: true
        }
    }
})
</script>