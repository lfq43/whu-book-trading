import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'
import Home from '../views/Home.vue'
import Publish from '../views/Publish.vue'
import MyBatches from '../views/MyBatches.vue'
import AdminPanel from '../views/AdminPanel.vue'
import { useUserStore } from '../stores/user'
import BatchDetail from "../views/BatchDetail.vue";
import UserSpace from "../views/UserSpace.vue";

const routes = [
    { path: '/', name: 'Home', component: Home },
    { path: '/login', name: 'Login', component: Login },
    { path: '/register', name: 'Register', component: Register },
    { path: '/publish', name: 'Publish', component: Publish, meta: { requiresAuth: true } },
    { path: '/my-batches', name: 'MyBatches', component: MyBatches, meta: { requiresAuth: true } },
    {
        path: '/batch/:id',
        name: 'BatchDetail',
        component: BatchDetail
    },
    {
        path: '/user/:id',
        name: 'UserSpace',
        component: UserSpace
    },    {
        path: '/admin',
        name: 'AdminPanel',
        component: AdminPanel,
        meta: { requiresAuth: true }
    }]

const router = createRouter({
    history: createWebHistory(),
    routes
})

// 路由守卫：需要登录的页面自动跳转
router.beforeEach((to, from, next) => {
    const userStore = useUserStore()
    if (to.meta.requiresAuth && !userStore.isLoggedIn) {
        next('/login')
    } else if (to.name === 'AdminPanel' && !userStore.isAdmin) {
        next('/')
    } else {
        next()
    }
})

export default router