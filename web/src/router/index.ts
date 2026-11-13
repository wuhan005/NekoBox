import {createRouter, createWebHistory, type RouteRecordRaw} from "vue-router";
import {useAuthStore} from "@/store";

const allRoutes: Array<RouteRecordRaw> = [
    {
        path: '/',
        name: 'home',
        component: () => import('@/pages/Home.vue'),
    },
    {path: '/login', name: 'login', redirect: '/sign-in'},
    {
        path: '/sign-in',
        name: 'sign-in',
        component: () => import('@/pages/auth/SignIn.vue'),
        meta: {
            signOutRequired: true,
        }
    },
    {path: '/register', name: 'register', redirect: '/sign-up'},
    {
        path: '/sign-up',
        name: 'sign-up',
        component: () => import('@/pages/auth/SignUp.vue'),
        meta: {
            signOutRequired: true,
        }
    },
    {
        path: '/forgot-password',
        name: 'forgot-password',
        component: () => import('@/pages/auth/ForgotPassword.vue'),
        meta: {
            signOutRequired: true,
        }
    },
    {
        path: '/recover-password',
        name: 'recover-password',
        component: () => import('@/pages/auth/RecoverPassword.vue'),
        meta: {
            signOutRequired: true,
        }
    },
    {
        path: '/_/:domain',
        name: 'profile',
        component: () => import('@/pages/user/Profile.vue'),
    },
    {
        path: '/_/:domain/:questionID',
        name: 'question',
        component: () => import('@/pages/user/QuestionItem.vue'),
    },
    {
        path: '/mine/questions',
        name: 'my-questions',
        component: () => import('@/pages/mine/QuestionList.vue'),
        meta: {
            signInRequired: true,
        }
    },
    {
        path: '/mine/deactivate',
        name: 'deactivate-account',
        component: () => import('@/pages/mine/DeactivateAccount.vue'),
        meta: {
            signInRequired: true,
        }
    },
    {
        path: '/settings',
        name: 'settings',
        component: () => import('@/pages/mine/Settings.vue'),
        meta: {
            signInRequired: true,
        }
    },

    {
        path: '/change-logs',
        name: 'change-logs',
        component: () => import('@/pages/general/ChangeLogs.vue'),
    },
    {
        path: '/pixel',
        name: 'pixel',
        component: () => import('@/pages/general/Pixel.vue'),
    },
    {
        path: '/sponsor',
        name: 'sponsor',
        component: () => import('@/pages/general/Sponsor.vue'),
    },

    {
        path: '/:pathMatch(.*)*',
        name: 'NotFound',
        component: () => import('@/pages/404.vue'),
    }
]

let questionListPosition: {left: number; top: number} | undefined

const router = createRouter({
    history: createWebHistory(),
    routes: allRoutes,
    scrollBehavior(to, from, savedPosition) {
        if (from.name === 'question' && to.name === 'my-questions') {
            const position = savedPosition ?? questionListPosition
            if (position) return {...position, behavior: 'instant'}
        }
        if (from.name === 'question' && to.name === 'profile' && savedPosition) {
            return {...savedPosition, behavior: 'instant'}
        }
        return {el: '#app', top: 0, behavior: 'smooth'}
    },
})

router.beforeEach((to, _, next) => {
    const authStore = useAuthStore()
    if (to.meta.signOutRequired && authStore.isSignedIn) {
        next({name: 'home'})
    } else if (to.meta.signInRequired && !authStore.isSignedIn) {
        next({name: 'sign-in', query: {to: to.fullPath}})
    } else {
        next()
    }
})

router.afterEach((to, from, failure) => {
    if (failure) return
    if (from.name === 'my-questions' && to.name === 'question') {
        questionListPosition = {left: window.scrollX, top: window.scrollY}
    } else if (to.name !== 'question' && to.name !== 'my-questions') {
        questionListPosition = undefined
    }
    document.title = to.meta.title ? `${to.meta.title} - NekoBox` : 'NekoBox'
})

export default router
