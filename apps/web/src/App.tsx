import { createBrowserRouter } from 'react-router-dom'
import { AppLayout } from '@/components/layout/AppLayout'
import { CaseCreate } from '@/pages/CaseCreate'
import { CaseDetail } from '@/pages/CaseDetail'
import { CaseList } from '@/pages/CaseList'
import { EstimateCreate } from '@/pages/cases/[caseId]/estimates/EstimateCreate'
import { EstimateDetail } from '@/pages/cases/[caseId]/estimates/EstimateDetail'
import { EstimateList } from '@/pages/cases/[caseId]/estimates/EstimateList'
import { Dashboard } from '@/pages/Dashboard'
import { CaseConversation } from '@/pages/cases/CaseConversation'
import { NotFound } from '@/pages/NotFound'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppLayout />,
    children: [
      { index: true, element: <Dashboard /> },
      { path: 'cases', element: <CaseList /> },
      { path: 'cases/new', element: <CaseCreate /> },
      { path: 'cases/:caseId', element: <CaseDetail /> },
      { path: 'cases/:caseId/conversation', element: <CaseConversation /> },
      { path: 'cases/:caseId/estimates', element: <EstimateList /> },
      { path: 'cases/:caseId/estimates/new', element: <EstimateCreate /> },
      {
        path: 'cases/:caseId/estimates/:estimateId',
        element: <EstimateDetail />,
      },
      { path: '*', element: <NotFound /> },
    ],
  },
])
