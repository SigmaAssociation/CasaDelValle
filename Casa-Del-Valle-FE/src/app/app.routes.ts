import { Routes } from '@angular/router';
import { StartPage } from './pages/start-page/start-page';
import { Registro } from './pages/registro/registro';
import { Login } from './pages/login/login';
import { Perfil } from './pages/perfil/perfil';
import { RegistrarCabana } from './pages/registrar-cabana/registrar-cabana';
import { MyCabinsPage } from './pages/my-cabins-page/my-cabins-page';
import { UserStoryGuidePage } from './pages/user-story-guide-page/user-story-guide-page';
import { ModelGuidePage } from './pages/model-guide-page/model-guide-page';
import { DesignGuidePage } from './pages/design-guide-page/design-guide-page';
import { BackendGuidePage } from './pages/backend-guide-page/backend-guide-page';
import { GuidePage } from './pages/guide-page/guide-page';
import { CabinEditPage } from './pages/cabin-edit-page/cabin-edit-page';
import { authGuard } from './guards/auth-guard';

export const routes: Routes = [
    {
        path: 'guide/user-story',
        component: UserStoryGuidePage
    },
    {
        path: 'guide/model',
        component: ModelGuidePage
    },
    {
        path: 'guide/design',
        component: DesignGuidePage
    },
    {
        path: 'guide/backend',
        component: BackendGuidePage
    },
    {
        path: 'guide',
        component: GuidePage
    },
    {
        path: '',
        component: StartPage,
    },
    {
        path: 'cabins/:id/edit',
        component: CabinEditPage,
        canActivate: [authGuard],
    },
    {
        path: 'mis-cabanas',
        component: MyCabinsPage,
        canActivate: [authGuard],
    },
    {
        path: 'registro', component: Registro
    },
    {
        path: 'login', component: Login
    },
    {
        path: 'perfil',
        component: Perfil,
        canActivate: [authGuard],
    },
    {
        path: 'registrar-cabana',
        component: RegistrarCabana,
        canActivate: [authGuard],
    },
    {
        path: '**',
        redirectTo: ''
    },

];
