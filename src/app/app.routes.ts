import { Routes } from '@angular/router';
import { LandingComponent } from './components/landing/landing.component';
import { authGuard } from './guards/auth.guard';

export const routes: Routes = [
  { path: '', component: LandingComponent },
  {
    path: 'dashboard',
    canActivate: [authGuard],
    loadComponent: () => import('./components/dashboard/dashboard.component').then(m => m.DashboardComponent)
  },
  {
    path: 'screener',
    canActivate: [authGuard],
    loadComponent: () => import('./components/screener/screener.component').then(m => m.ScreenerComponent)
  },
  {
    path: 'alerts',
    canActivate: [authGuard],
    loadComponent: () => import('./components/alerts/alerts.component').then(m => m.AlertsComponent)
  },
  {
    path: 'login',
    loadComponent: () => import('./components/auth/login/login.component').then(m => m.LoginComponent)
  },
  {
    path: 'register',
    loadComponent: () => import('./components/auth/register/register.component').then(m => m.RegisterComponent)
  },
  { path: '**', redirectTo: '' }
];
