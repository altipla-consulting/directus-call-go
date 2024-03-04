
<template>
  <VSelect
    v-model="model"
    :items="functions ?? []"
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
import { computed, inject, ref, watch, type Ref, nextTick } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { installVueQuery } from '../global'
import { useApi } from '@directus/extensions-sdk'
import { firstBy } from 'thenby'
import type { Server } from '../models/servers'

installVueQuery()

let props = defineProps({
  value: {
    type: String,
    default: null,
  },
})
let emit = defineEmits(['input', 'setFieldValue'])

let values = inject<Ref<{ server: null | string }>>('values')

let model = ref<string | null>(props.value ? `${values!.value.server}||${props.value}` : null)
watch(model, () => {
  if (model.value) {
    let val = model.value.split('||')
    emit('input', val[1])
    void nextTick(() => emit('setFieldValue', { field: 'server', value: val[0] }))
  } else {
    emit('input', null)
    void nextTick(() => emit('setFieldValue', { field: 'server', value: null }))
  }
})

let api = useApi()

let { data: servers, isLoading: isLoadingServers, isError: isErrorServers, error: errorServers } = useQuery({
  queryKey: ['servers'],
  queryFn: async () => {
    let reply = await api.get<Server[]>('/call-go/servers')
    return reply.data
  },
})

let { data: functions, isLoading: isLoadingFunctions, isError: isErrorFunctions, error: errorFunctions } = useQuery({
  queryKey: ['functions', servers.value?.join(',')],
  enabled: computed<boolean>(() => !!servers.value?.length),
  retry: 3,
  queryFn: async () => {
    let data = await Promise.all(servers.value!.map(async server => {
      try {
        let reply = await api.get<string[]>(`/call-go/functions?server=${server.alias}`)
        return reply.data.map((fnname: string) => {
          return {
            text: servers.value!.length > 1 ? `${server.alias} / ${fnname}` : fnname,
            value: `${server.alias}||${fnname}`,
          }
        })
      } catch (error) {
        let err = new Error(`Cannot load functions from server ${server.alias}`)
        err.cause = error
        throw err
      }
    }))

    let all = data.flat()
    all.sort(firstBy('text'))
    return all
  },
})

let isLoading = computed(() => isLoadingServers.value || isLoadingFunctions.value)
let isError = computed(() => isErrorServers.value || isErrorFunctions.value)
let error = computed(() => errorServers.value || errorFunctions.value)
</script>

<style type="text-css">
.notice {
	margin-top: 16px;
}
</style>
