import { Routes } from '@angular/router';
import { StartPage } from './pages/start-page/start-page';
import { Registro } from './pages/registro/registro';
import { Login } from './pages/login/login';
import { Perfil } from './pages/perfil/perfil';
import { RegistrarCabana } from './pages/registrar-cabana/registrar-cabana';
import { UserStoryGuidePage } from './pages/user-story-guide-page/user-story-guide-page';
import { ModelGuidePage } from './pages/model-guide-page/model-guide-page';
import { DesignGuidePage } from './pages/design-guide-page/design-guide-page';
import { BackendGuidePage } from './pages/backend-guide-page/backend-guide-page';
import { GuidePage } from './pages/guide-page/guide-page';
import { CabinEditPage } from './pages/cabin-edit-page/cabin-edit-page';
import { CabinListPage } from './pages/cabin-list-page/cabin-list-page';

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
        component: CabinListPage,
    },    {
        path: 'cabins/:id/edit',
        component: CabinEditPage,
    },

    {
        path: 'registro', component: Registro
    },
    {
        path: 'login', component: Login
    },
    {
        path: 'perfil', component: Perfil
    },
    {
        path: 'registrar-cabana', component: RegistrarCabana
    },
    {
        path: '**',
        redirectTo: ''
    },

];
