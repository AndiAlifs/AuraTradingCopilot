import { Component, EventEmitter, Output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-trading-profile-modal',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './trading-profile-modal.component.html',
  styleUrl: './trading-profile-modal.component.css'
})
export class TradingProfileModalComponent {
  @Output() close = new EventEmitter<void>();
  profileForm: FormGroup;
  isSubmitting = false;

  constructor(private fb: FormBuilder, private authService: AuthService) {
    this.profileForm = this.fb.group({
      riskTolerance: ['', Validators.required],
      preferredStrategy: ['', Validators.required]
    });
  }

  onSubmit() {
    if (this.profileForm.valid) {
      this.isSubmitting = true;
      this.authService.updateProfile(this.profileForm.value).subscribe({
        next: () => {
          this.isSubmitting = false;
          this.close.emit();
        },
        error: (err) => {
          console.error('Failed to update profile', err);
          this.isSubmitting = false;
        }
      });
    }
  }

  onCancel() {
    this.close.emit();
  }
}
