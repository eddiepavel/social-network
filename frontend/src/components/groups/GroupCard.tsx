import { useNavigate } from 'react-router-dom';
import type { Group } from '../../types';

interface GroupCardProps {
  group: Group;
}

export function GroupCard({ group }: GroupCardProps) {
  const navigate = useNavigate();

  const handleClick = () => {
    navigate(`/groups/${group.group_id}`);
  };

  return (
    <div
      className="cursor-pointer rounded-2xl border border-white/10 bg-card/70 p-6 shadow-soft hover:border-primary-400/40 transition"
      onClick={handleClick}
    >
      <div className="flex items-start justify-between gap-3">
        <h3 className="text-lg font-semibold text-white">{group.group_name}</h3>
        {group.member_count !== undefined && (
          <span className="rounded-full bg-white/5 px-3 py-1 text-xs text-muted">
            {group.member_count} members
          </span>
        )}
      </div>
      <p className="mt-3 text-sm text-muted line-clamp-3">{group.description}</p>
      <div className="mt-4 flex items-center justify-between text-xs text-muted">
        <span>
          Created {new Date(group.created_at).toLocaleDateString()}
        </span>
      </div>
    </div>
  );
}
