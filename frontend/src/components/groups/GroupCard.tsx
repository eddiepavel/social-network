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
    <div className="group-card" onClick={handleClick}>
      <div className="group-card-header">
        <h3>{group.group_name}</h3>
        {group.member_count !== undefined && (
          <span className="member-count">{group.member_count} members</span>
        )}
      </div>
      <p className="group-description">{group.description}</p>
      <div className="group-card-footer">
        <span className="created-date">
          Created {new Date(group.created_at).toLocaleDateString()}
        </span>
      </div>
    </div>
  );
}
