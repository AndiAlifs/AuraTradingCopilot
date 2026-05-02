import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { DashboardComponent } from '../dashboard/dashboard.component';

@Component({
  selector: 'app-landing',
  standalone: true,
  imports: [RouterLink, DashboardComponent],
  templateUrl: './landing.component.html',
  styleUrl: './landing.component.css'
})
export class LandingComponent {

}
