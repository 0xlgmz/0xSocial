import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { AuthenticatedLayout } from '../../components/layout/AuthenticatedLayout'

type PlaceholderPageProps = {
  route: '/notifications'
  title: string
  description: string
  icon: IconDefinition
}

export function PlaceholderPage({ route, title, description, icon }: PlaceholderPageProps) {
  return <AuthenticatedLayout title={title} route={route}><section className="card min-h-[calc(100vh-8rem)] border border-base-300 bg-base-100 shadow-sm"><div className="card-body items-center justify-center text-center"><div className="grid size-20 place-items-center rounded-full bg-primary/10 text-3xl text-primary"><FontAwesomeIcon icon={icon}/></div><span className="badge badge-outline badge-primary mt-3">Coming soon</span><h2 className="mt-2 text-3xl font-semibold tracking-[-.035em]">{title}</h2><p className="max-w-sm text-sm leading-6 text-base-content/55">{description}</p></div></section></AuthenticatedLayout>
}
