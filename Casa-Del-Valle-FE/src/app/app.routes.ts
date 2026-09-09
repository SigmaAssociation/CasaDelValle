import { Routes } from '@angular/router';
import { StartPage } from './pages/start-page/start-page';
import { UserStoryGuidePage } from './pages/user-story-guide-page/user-story-guide-page';
import { ModelGuidePage } from './pages/model-guide-page/model-guide-page';
import { DesignGuidePage } from './pages/design-guide-page/design-guide-page';
import { BackendGuidePage } from './pages/backend-guide-page/backend-guide-page';
import { GuidePage } from './pages/guide-page/guide-page';

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
        path: '**',
        redirectTo: ''
    }
];
