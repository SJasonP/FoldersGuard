import { useEffect, useMemo, useState } from 'react';
import type { HookAPI as ModalHookAPI } from 'antd/es/modal/useModal';
import { ListLocalProjects } from '../../wailsjs/go/main/App';
import type { LocalProjectRow, LocalProjectSummary } from '../types';
import { formatDateTime } from '../formatters';
import { showOperationError } from '../components/common/operationError';

type UseLocalProjectsArgs = {
  language: 'en-US' | 'zh-CN';
  modalApi: ModalHookAPI;
  t: (key: string) => string;
};

export function useLocalProjects({ language, modalApi, t }: UseLocalProjectsArgs) {
  const [projectSearch, setProjectSearch] = useState('');
  const [projects, setProjects] = useState<LocalProjectSummary[]>([]);
  const [projectsLoading, setProjectsLoading] = useState(true);
  const [projectsError, setProjectsError] = useState<string | null>(null);
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null);

  const loadProjects = async () => {
    setProjectsLoading(true);
    setProjectsError(null);
    try {
      const nextProjects = await ListLocalProjects();
      setProjects(nextProjects);
    } catch (error) {
      setProjects([]);
      setProjectsError(t('errorLoadingProjects'));
      showOperationError(modalApi, t('errorLoadingProjects'), error, t);
    } finally {
      setProjectsLoading(false);
    }
  };

  useEffect(() => {
    void loadProjects();
  }, []);

  const visibleProjects = useMemo<LocalProjectRow[]>(
    () =>
      projects
        .filter((project) => {
          const query = projectSearch.trim().toLowerCase();
          if (query === '') {
            return true;
          }
          return (
            project.projectId.toLowerCase().includes(query) ||
            project.fileName.toLowerCase().includes(query) ||
            project.availabilityStatus.toLowerCase().includes(query)
          );
        })
        .map((project) => {
          const modifiedAt = project.modifiedAt ? new Date(project.modifiedAt) : null;
          return {
            key: project.projectId,
            projectId: project.projectId,
            fileName: project.fileName,
            modifiedTime: formatDateTime(modifiedAt),
            modifiedAtMs: modifiedAt?.getTime() ?? 0,
            availabilityStatus: t(project.availabilityStatus),
          };
        }),
    [language, projectSearch, projects, t],
  );

  const selectedProject = useMemo(
    () => projects.find((project) => project.projectId === selectedProjectId) ?? null,
    [projects, selectedProjectId],
  );

  return {
    projectSearch,
    setProjectSearch,
    projectsLoading,
    projectsError,
    selectedProject,
    selectedProjectId,
    setSelectedProjectId,
    visibleProjects,
    loadProjects,
  };
}
