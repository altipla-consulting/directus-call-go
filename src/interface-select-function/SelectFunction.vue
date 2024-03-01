
<template>
	<VSelect
		v-model="model"
		:items="functions"
		placeholder="Function Name"
		search
		:disabled="isLoading || isError"
	/>
	<VNotice type="danger" v-if="isError" class="notice">
		{{error}}<br>
		{{error?.cause}}
	</VNotice>
</template>

<script lang="ts" setup>
import { computed, ref, toRef, watch } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { installVueQuery } from '../global'
import { useApi } from '@directus/extensions-sdk'
import { firstBy } from 'thenby'

installVueQuery()

let props = defineProps({
	value: {
		type: String,
		default: null,
	},
})
let emit = defineEmits(['input'])

let model = ref<string | null>(props.value)
watch(toRef(props, 'value'), () => {
	model.value = props.value
})
watch(model, () => {
	emit('input', model.value)
})

let api = useApi()

let servers = ref<string[]>([])

const { data, isLoading, isError, error } = useQuery({
	queryKey: ['functions', servers.value.join(',')],
	enabled: computed<boolean>(() => !!servers.value.length),
	retry: 3,
  queryFn: () => {
		return Promise.all(servers.value.map(async (server) => {
			try {
				let reply = await api.get(`/call-go/functions?server=${server}`)
				return reply.data.map((fn: string) => ({
					text: fn,
					value: [server, fn].join('||'),
				}))
			} catch (error) {
				let err = new Error(`Cannot load functions from server ${server}`)
				err.cause = error
				throw err
			}
		}))
	},
})
let functions = computed(() => {
	if (data.value) {
		let all = data.value.flat()
		all.sort(firstBy('text'))
		return all
	}
	return []
})

async function load() {
	let reply = await api.get('/call-go/servers')
	servers.value = reply.data
}
load()
</script>

<style type="text-css">
.notice {
	margin-top: 16px;
}
</style>
