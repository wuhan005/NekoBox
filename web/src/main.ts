import {createApp} from 'vue'
import App from './App.vue'
import './style.less'
import store from './store'
import router from './router'
import '@/api/interceptor.ts'
import 'vue3-toastify/dist/index.css'

import {configure, defineRule} from "vee-validate";
import {localize, setLocale} from "@vee-validate/i18n";
import {all as AllRules} from '@vee-validate/rules'
import zh_CN from "@vee-validate/i18n/dist/locale/zh_CN.json";

import {VueReCaptcha} from 'vue-recaptcha-v3'

import UIkit from 'uikit';
import Icons from 'uikit/dist/js/uikit-icons';
import "../node_modules/uikit/dist/js/uikit-icons.min.js"

import "vue-loading-skeleton/dist/style.css"

UIkit.use(Icons);

Object.keys(AllRules).forEach(rule => {
    defineRule(rule, AllRules[rule])
})
configure({
    generateMessage: localize({
        zh_CN,
    }),
});
setLocale('zh_CN')

const app = createApp(App)
app.use(store)
app.use(router)

// Register reCAPTCHA only when site_key is configured; in go_captcha mode it can be left empty.
const recaptchaSiteKey = import.meta.env.VITE_RECAPTCHA_SITE_KEY
if (recaptchaSiteKey) {
    app.use(VueReCaptcha, {
        siteKey: recaptchaSiteKey,
        loaderOptions: {
            useRecaptchaNet: true,
            autoHideBadge: true,
        }
    })
}

app.mount('#app')