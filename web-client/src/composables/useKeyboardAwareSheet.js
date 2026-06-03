import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'

export function useKeyboardAwareSheet() {
  const keyboardHeight = ref(0)
  const focusedFieldKey = ref('')
  const activeFieldAnchor = ref('')

  const isKeyboardVisible = computed(() => Number(keyboardHeight.value || 0) > 0)

  const safeFieldId = (key = '') => {
    return `form-field-${String(key).replace(/[^a-zA-Z0-9_-]/g, '-')}`
  }

  const getFieldKey = (field) => {
    if (typeof field === 'string') return field
    return field?.key || field?.name || ''
  }

  const fieldDomId = (field) => {
    return safeFieldId(getFieldKey(field))
  }

  const formSheetStyle = computed(() => ({
    bottom: '0px',
    height: '88vh',
    maxHeight: '88vh',
  }))

  const formScrollContentStyle = computed(() => {
    const extra = Number(keyboardHeight.value || 0)
    return {
      paddingBottom: `${extra + 120}px`,
    }
  })

  const scrollToField = (field) => {
    const key = getFieldKey(field)
    if (!key) return

    const id = safeFieldId(key)
    activeFieldAnchor.value = ''

    nextTick(() => {
      setTimeout(() => {
        activeFieldAnchor.value = id
      }, 300)
    })
  }

  const handleFieldFocus = (field) => {
    focusedFieldKey.value = getFieldKey(field)
  }

  const handleFieldBlur = (field, validator) => {
    const key = getFieldKey(field)

    if (typeof validator === 'function') {
      validator(field)
    }

    if (focusedFieldKey.value === key) {
      focusedFieldKey.value = ''
    }

    setTimeout(() => {
      if (!focusedFieldKey.value) {
        keyboardHeight.value = 0
        activeFieldAnchor.value = ''
      }
    }, 180)
  }

  const handleKeyboardHeightChange = (e) => {
    const height = Number(e?.detail?.height || 0)
    keyboardHeight.value = Math.max(0, height)

    if (height > 0 && focusedFieldKey.value) {
      scrollToField(focusedFieldKey.value)
    }

    if (height <= 0) {
      activeFieldAnchor.value = ''
    }
  }

  const handleGlobalKeyboardHeightChange = (res) => {
    const height = Number(res?.height || 0)
    keyboardHeight.value = Math.max(0, height)

    if (height > 0 && focusedFieldKey.value) {
      scrollToField(focusedFieldKey.value)
    }

    if (height <= 0) {
      focusedFieldKey.value = ''
      activeFieldAnchor.value = ''
    }
  }

  const hideKeyboard = () => {
    focusedFieldKey.value = ''
    activeFieldAnchor.value = ''
    keyboardHeight.value = 0

    const uniApi = typeof uni !== 'undefined' ? uni : null
    if (uniApi?.hideKeyboard) {
      uniApi.hideKeyboard()
    }
  }

  const resetKeyboardSheet = () => {
    hideKeyboard()
  }

  onMounted(() => {
    const uniApi = typeof uni !== 'undefined' ? uni : null
    if (uniApi?.onKeyboardHeightChange) {
      uniApi.onKeyboardHeightChange(handleGlobalKeyboardHeightChange)
    }
  })

  onUnmounted(() => {
    const uniApi = typeof uni !== 'undefined' ? uni : null
    if (uniApi?.offKeyboardHeightChange) {
      uniApi.offKeyboardHeightChange(handleGlobalKeyboardHeightChange)
    }
  })

  return {
    keyboardHeight,
    focusedFieldKey,
    activeFieldAnchor,
    isKeyboardVisible,
    formSheetStyle,
    formScrollContentStyle,
    fieldDomId,
    handleFieldFocus,
    handleFieldBlur,
    handleKeyboardHeightChange,
    hideKeyboard,
    resetKeyboardSheet,
  }
}
