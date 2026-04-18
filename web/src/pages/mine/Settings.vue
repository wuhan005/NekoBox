<template>
  <UkTabs
      v-model="currentTab"
      :tabs="TABS"
      @change="handleChangeTab"
  >
    <template #profile>
      <Form @submit="updateProfile">
        <div class="uk-margin">
          <label class="uk-form-label" for="email">电子邮箱</label>
          <input v-model="profile.email" id="email" class="uk-input" type="text" disabled/>
        </div>
        <div class="uk-margin">
          <label class="uk-form-label" for="name">昵称</label>
          <Field v-model="updateProfileForm.name" id="name" name="name" class="uk-input" type="text" rules="required"
                 label="昵称"/>
          <ErrorMessage class="field-error-message" name="name"/>
        </div>
        <div class="uk-margin">
          <label class="uk-form-label" for="oldPassword">旧密码</label>
          <Field v-model="updateProfileForm.oldPassword" class="uk-input" id="oldPassword" name="oldPassword"
                 type="password"
                 placeholder="修改密码使用，留空则不修改"/>
          <ErrorMessage class="field-error-message" name="oldPassword"/>
        </div>
        <div class="uk-margin">
          <label class="uk-form-label" for="newPassword">新密码</label>
          <Field v-model="updateProfileForm.newPassword" class="uk-input" id="newPassword" name="newPassword"
                 type="password"
                 placeholder="修改密码使用，留空则不修改"/>
          <ErrorMessage class="field-error-message" name="newPassword"/>
        </div>

        <div class="uk-margin">
          <button type="submit" class="uk-button uk-button-primary">修改信息</button>
          <button type="button" class="uk-button uk-button-danger" @click="handleSignOut">退出登录</button>
        </div>
      </Form>
    </template>

    <template #box>
      <Form @submit="updateBoxSettings">
        <div class="uk-margin">
          <label class="uk-form-label" for="intro">提问箱介绍</label>
          <input v-model="updateMineBoxSettingsForm.intro" id="intro" name="intro" class="uk-input" type="text">
        </div>
        <div class="uk-margin">
          <label class="uk-form-label" for="form-stacked-text">新提问通知方式</label>
          <br/>
          <label class="uk-text-small">
            <input v-model="notifyTypeEmail" name="notifyType" type="checkbox"
                   class="uk-checkbox"/> 邮件
          </label>
        </div>
        <div class="uk-margin">
          <label class="uk-form-label" for="form-stacked-text">提问箱头像</label>
          <br/>
          <div uk-form-custom="target: true">
            <input ref="avatarImageUploader" name="avatar" type="file" accept="image/*"
                   @change="handleSelectAvatarImage">
            <input class="uk-input uk-form-width-large" type="text" placeholder="点击选择个人提问箱头像" disabled>
          </div>
        </div>
        <div class="uk-margin">
          <label class="uk-form-label" for="form-stacked-text">提问箱背景</label>
          <br/>
          <div uk-form-custom="target: true">
            <input ref="backgroundImageUploader" name="avatar" type="file" accept="image/*"
                   @change="handleSelectBackgroundImage">
            <input class="uk-input uk-form-width-large" type="text" placeholder="点击选择提问箱背景" disabled>
          </div>
        </div>
        <div class="uk-margin">
          <button type="submit" class="uk-button uk-button-primary" :disabled="boxSettingsLoading">
            {{ boxSettingsLoading ? '保存中...' : '保存配置' }}
          </button>
        </div>
      </Form>
    </template>

    <template #harassment>
      <div class="uk-margin">
        <Form @submit="updateHarassmentSettings">
          <div class="uk-margin">
            <label class="uk-form-label" for="form-stacked-text">提问限制</label>
            <br/>
            <label class="uk-text-small">
              <input v-model="harassmentSettingTypeRegisterOnly" name="notifyType" type="checkbox"
                     class="uk-checkbox"/> 仅允许注册用户向我提问
            </label>
          </div>
          <div class="uk-margin">
            <label class="uk-form-label">屏蔽词设置（使用半角逗号 <code>,</code> 分隔，最多支持 10 个屏蔽词，每个屏蔽词最大长度为
              10）</label>
            <input v-model="updateMineHarassmentSettingForm.blockWords" name="blockWords" class="uk-input" type="text"/>
          </div>
          <div class="uk-margin">
            <button type="submit" class="uk-button uk-button-primary">更新防骚扰设置</button>
          </div>
        </Form>
      </div>
    </template>

    <template #account>

      <dl class="uk-description-list uk-description-list-divider">
        <dt>
          <h2 class="uk-legend">数据导出</h2>
          <span
              class="uk-text-muted">您可以导出您在 NekoBox 中的所有个人数据，包括你的基本信息、收到的问题以及回答。
            </span>
          <br><br>
          <button class="uk-button uk-button-default" @click="handleExportData" :disabled="exportDataLoading">
            {{ exportDataLoading ? '导出中...' : '导出我的数据' }}
          </button>
        </dt>
        <dt>
          <h2 class="uk-legend">账号停用</h2>
          <span class="uk-text-muted">您随时可以选择停用您的账号。停用后，您的账号将无法登录，您的提问箱页面以及提问将无法访问，其他人也无法再给您发送新的提问。<b>该操作无法撤销！请谨慎操作！</b></span>
          <br><br>
          <button class="uk-button uk-button-danger" @click="handleDeactivateAccount">停用我的账号</button>
        </dt>
      </dl>
    </template>
  </UkTabs>
</template>

<script setup lang="ts">
import {ref, onMounted} from "vue";
import UkTabs from "@/components/UkTabs.vue";
import {
  getMineBoxSettings,
  getMineProfile,
  updateMineProfile,
  type MineBoxSettings,
  type MineProfile,
  type UpdateMineBoxSettingsRequest,
  type UpdateMineProfileRequest, updateMineBoxSettings, type UpdateMineHarassmentSettingsRequest,
  getMineHarassmentSettings, updateMineHarassmentSettings, exportData
} from "@/api/mine.ts";
import {Form, Field, ErrorMessage} from 'vee-validate';
import {ToastSuccess} from "@/utils/notify.ts";
import {useAuthStore} from "@/store";
import {useRouter} from "vue-router";
import {saveAs} from 'file-saver';

const router = useRouter()
const authStore = useAuthStore();

const TABS = [
  {name: 'profile', label: '个人信息'},
  {name: 'box', label: '提问箱设置'},
  {name: 'harassment', label: '防骚扰设置'},
  {name: 'account', label: '账号设置'},
]
const currentTab = ref<string>('profile')
const handleChangeTab = (tab: string) => {
  currentTab.value = tab
}

// ===== PROFILE SETTINGS ===== //
const profile = ref<MineProfile>({
  name: '',
  email: '',
} as MineProfile)
const updateProfileForm = ref<UpdateMineProfileRequest>({
  name: '',
  oldPassword: '',
  newPassword: '',
})
const fetchProfile = () => {
  getMineProfile().then(res => {
    profile.value = res
    updateProfileForm.value.name = res.name
    authStore.setProfileName(res.name)
  })
}
const updateProfile = () => {
  updateMineProfile(updateProfileForm.value).then(res => {
    ToastSuccess(res)
  }).finally(() => {
    updateProfileForm.value.oldPassword = ''
    updateProfileForm.value.newPassword = ''
    fetchProfile()
  })
}
const handleSignOut = () => {
  authStore.signOut()
  ToastSuccess('账号已退出登录')
  router.push({
    name: 'home'
  })
}

// ===== BOX SETTINGS ===== //
const boxSettings = ref<MineBoxSettings>({
  intro: '',
  notifyType: 'none',
  avatarURL: '',
  backgroundURL: '',
} as MineBoxSettings)
const notifyTypeEmail = ref<boolean>(false)
const updateMineBoxSettingsForm = ref<UpdateMineBoxSettingsRequest>({
  intro: '',
  notifyType: 'none',
  avatar: null,
  background: null,
})
const boxSettingsLoading = ref<boolean>(false)
const avatarImageUploader = ref<HTMLInputElement | null>(null)
const backgroundImageUploader = ref<HTMLInputElement | null>(null)
const fetchBoxSettings = () => {
  getMineBoxSettings().then(res => {
    boxSettings.value = res
    updateMineBoxSettingsForm.value.intro = res.intro
    notifyTypeEmail.value = res.notifyType === 'email' // TODO: Find a better way to handle this
  })
}
const updateBoxSettings = () => {
  if (notifyTypeEmail.value) {
    updateMineBoxSettingsForm.value.notifyType = 'email'
  } else {
    updateMineBoxSettingsForm.value.notifyType = 'none'
  }

  boxSettingsLoading.value = true
  updateMineBoxSettings(updateMineBoxSettingsForm.value).then(res => {
    ToastSuccess(res)
  }).finally(() => {
    boxSettingsLoading.value = false

    updateMineBoxSettingsForm.value.avatar = null
    updateMineBoxSettingsForm.value.background = null
    if (avatarImageUploader.value) {
      avatarImageUploader.value.value = ''
    }
    if (backgroundImageUploader.value) {
      backgroundImageUploader.value.value = ''
    }
    fetchBoxSettings()
  })
}
const handleSelectAvatarImage = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    updateMineBoxSettingsForm.value.avatar = target.files[0]
  } else {
    updateMineBoxSettingsForm.value.avatar = null
  }
}
const handleSelectBackgroundImage = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    updateMineBoxSettingsForm.value.background = target.files[0]
  } else {
    updateMineBoxSettingsForm.value.background = null
  }
}

// ===== HARASSMENT SETTINGS ===== //
const updateMineHarassmentSettingForm = ref<UpdateMineHarassmentSettingsRequest>({
  harassmentSettingType: 'none',
  blockWords: '',
} as UpdateMineHarassmentSettingsRequest)
const harassmentSettingTypeRegisterOnly = ref<boolean>(false)
const fetchHarassmentSettings = () => {
  getMineHarassmentSettings().then(res => {
    harassmentSettingTypeRegisterOnly.value = res.harassmentSettingType === 'register_only'
    updateMineHarassmentSettingForm.value.harassmentSettingType = res.harassmentSettingType
    updateMineHarassmentSettingForm.value.blockWords = res.blockWords
  })
}
const updateHarassmentSettings = () => {
  if (harassmentSettingTypeRegisterOnly.value) {
    updateMineHarassmentSettingForm.value.harassmentSettingType = 'register_only'
  } else {
    updateMineHarassmentSettingForm.value.harassmentSettingType = 'none'
  }

  updateMineHarassmentSettings(updateMineHarassmentSettingForm.value).then(res => {
    ToastSuccess(res)
  }).finally(() => {
    fetchHarassmentSettings()
  })
}

const exportDataLoading = ref<boolean>(false)
const handleExportData = () => {
  exportDataLoading.value = true
  exportData().then(res => {
    saveAs(res)
  }).finally(() => {
    exportDataLoading.value = false
  })
}

const handleDeactivateAccount = () => {
  router.push({
    name: 'deactivate-account'
  })
}

onMounted(() => {
  fetchProfile()
  fetchBoxSettings()
  fetchHarassmentSettings()
})
</script>

<style scoped>

</style>