export interface LabCatalogAttempt {
  id: string;
  status: string;
}

export interface LabCatalogItem {
  id: number;
  name: string;
  description?: string;
  collaboration: number;
  repository: string;
  attempt?: LabCatalogAttempt;
}

export interface LabCatalogListResponse {
  labs: LabCatalogItem[];
}

export interface LabCatalogStartResponse {
  attemptId: string;
  status: string;
  nextUrl: string;
}
