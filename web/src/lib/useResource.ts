import { ref, shallowRef } from "vue";
import { errorText } from "./notify";

export function useResource<T>(load: () => Promise<T>) {
  const data = shallowRef<T>();
  const error = ref<string>();
  const loading = ref(false);

  async function refresh() {
    loading.value = true;
    try {
      data.value = await load();
      error.value = undefined;
    } catch (e) {
      error.value = errorText(e);
    } finally {
      loading.value = false;
    }
  }

  void refresh();

  return { data, error, loading, refresh };
}
