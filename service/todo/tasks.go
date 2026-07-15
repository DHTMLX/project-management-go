package todo

import (
	"errors"
	uCtx "project-manager-go/api/context"
	"project-manager-go/common"
	"project-manager-go/data"
	"project-manager-go/data/todo"
	"project-manager-go/service"
	"time"
)

type tasks struct {
	tree  *service.TreeService
	store *todo.TodoStore
}

func (s *tasks) GetAll(userCtx uCtx.UserContext, dbCtx *data.DBContext) (arr []Task, err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	list := data.NewItemsList[Task]()
	err = s.store.Tasks.GetAll(dbCtx, list)

	return list.GetArray(), err
}

func (s *tasks) GetByProject(userCtx uCtx.UserContext, dbCtx *data.DBContext, projectId int) (arr []Task, err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	list := data.NewItemsList[Task]()
	err = s.store.Tasks.GetByProject(dbCtx, projectId, list)

	return list.GetArray(), err
}

func (s *tasks) Add(userCtx uCtx.UserContext, dbCtx *data.DBContext, op AddTask) (id int, err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	now := time.Now()
	task := op.Task
	task.CreationDate = &now

	id, err = s.store.Tasks.Add(dbCtx, &task)
	target := int(op.TargetID)
	project := int(op.ProjectID)
	parent := int(op.ParentID)

	var index int
	if target == parent && target != 0 {
		// add sub-task, push to the end of the branch
		index, err = s.store.Tasks.MaxBranchIndex(dbCtx, project, parent)
		if err != nil {
			return 0, err
		}
		index++
	} else if target > 0 {
		// add after target
		index, err = s.getMoveIndex(dbCtx, id, target, parent, op.Reverse)
		if err != nil {
			return 0, err
		}
	}

	if err != nil {
		return 0, err
	}

	err = s.tree.Move(dbCtx, id, project, parent, index)

	return id, err
}

func (s *tasks) Update(userCtx uCtx.UserContext, dbCtx *data.DBContext, id int, op UpdateTask) (err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	task := Task{}
	err = s.store.Tasks.GetOne(dbCtx, id, &task)
	if err != nil {
		return err
	}

	t := op.Task

	if op.ID == 0 {
		// only due_date
		task.DueDate = op.DueDate
		t = task
	}

	t.Index = task.Index

	err = s.store.Tasks.DeleteAssociations(dbCtx, id)
	if err != nil {
		return err
	}

	err = s.store.Tasks.Update(dbCtx, id, &t)

	return err
}

func (s *tasks) Delete(userCtx uCtx.UserContext, dbCtx *data.DBContext, id int) (children []int, err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	children, err = s.store.Tasks.DeleteCascade(dbCtx, id)

	return children, err
}

func (s *tasks) Clone(userCtx uCtx.UserContext, dbCtx *data.DBContext, op *CloneTask) (pull map[string]int, err error) {
	if op == nil || len(op.Batch) == 0 {
		return nil, nil
	}
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	targetTask := Task{}
	err = s.store.Tasks.GetOne(dbCtx, int(op.TargetID), &targetTask)
	if err != nil {
		return nil, err
	}

	index := targetTask.Index + 1
	err = s.store.Tasks.ShiftIndex(dbCtx, int(targetTask.ProjectID), int(targetTask.ParentID), index, 1)
	if err != nil {
		return nil, err
	}

	pull = make(map[string]int)
	root := op.Batch[0]
	id, err := s.createTempTask(dbCtx, root, int(op.ParentID), index)
	if err != nil {
		return nil, err
	}
	index++
	pull[root.ID] = id

	if len(op.Batch) > 1 {
		indices := make(map[string]int)
		for _, t := range op.Batch[1:] {
			parent := string(t.ParentID)
			id, err := s.createTempTask(dbCtx, t, pull[parent], indices[parent])
			if err != nil {
				return nil, err
			}
			pull[t.ID] = id
			indices[parent]++
		}
	}

	return pull, err
}

func (s *tasks) Move(userCtx uCtx.UserContext, dbCtx *data.DBContext, id int, op MoveTask) (moveId int, err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	if op.ID == 0 && len(op.IDs) == 0 {
		return 0, errors.New("invalid move params")
	}

	moveId = int(op.ID)
	if moveId == 0 {
		moveId = op.IDs[0]
	}

	task := Task{}
	err = s.store.Tasks.GetOne(dbCtx, moveId, &task)

	newProject := int(op.ProjectID)
	project := int(task.ProjectID)
	parent := int(op.ParentID)
	target := int(op.TargetID)

	switch op.Operation {
	case "project":
		err = s.ChangeProject(userCtx, dbCtx, moveId, newProject)
	case "indent":
		err = s.Indent(userCtx, dbCtx, moveId, project, parent)
	case "unindent":
		err = s.Unindent(userCtx, dbCtx, moveId, project, parent)
	default:
		err = s.ChangePosition(userCtx, dbCtx, moveId, target, project, parent, op.Reverse)
	}

	return moveId, err
}

func (s *tasks) ChangeProject(userCtx uCtx.UserContext, dbCtx *data.DBContext, id, projectId int) (err error) {
	return s.tree.ChangeNodeProject(dbCtx, id, projectId, nil)
}

func (s *tasks) ChangePosition(userCtx uCtx.UserContext, dbCtx *data.DBContext, id, targetId, projectId, parentId int, reverse bool) (err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	index, err := s.getMoveIndex(dbCtx, id, targetId, parentId, reverse)
	if err != nil {
		return err
	}

	err = s.tree.Move(dbCtx, id, projectId, parentId, index)

	return err
}

func (s *tasks) Indent(userCtx uCtx.UserContext, dbCtx *data.DBContext, id, projectId, parentId int) (err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	maxBranchIndex, err := s.store.Tasks.MaxBranchIndex(dbCtx, projectId, parentId)
	if err != nil {
		return err
	}
	err = s.tree.Move(dbCtx, id, projectId, parentId, maxBranchIndex+1)

	return err
}

func (s *tasks) Unindent(userCtx uCtx.UserContext, dbCtx *data.DBContext, id, projectId, parentId int) (err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	t := Task{}
	err = s.store.Tasks.GetOne(dbCtx, id, &t)
	if err != nil {
		return err
	}

	oldParentId := int(t.ParentID)
	oldParentTask := Task{}
	err = s.store.Tasks.GetOne(dbCtx, oldParentId, &oldParentTask)
	if err != nil {
		return err
	}

	err = s.tree.Move(dbCtx, id, projectId, parentId, oldParentTask.Index+1)

	return err
}

func (s *tasks) getMoveIndex(dbCtx *data.DBContext, id, targetId, parentId int, reverse bool) (index int, err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	task := Task{}
	err = s.store.Tasks.GetOne(dbCtx, id, &task)
	if err != nil {
		return 0, err
	}

	targetTask := Task{}
	err = s.store.Tasks.GetOne(dbCtx, targetId, &targetTask)
	if err != nil {
		return 0, err
	}
	project := int(targetTask.ProjectID)
	sameBranch := int(task.ParentID) == parentId

	if targetId == parentId {
		// if the targetId same with parentId, then move the task to the end of the branch
		index, err = s.store.Tasks.MaxBranchIndex(dbCtx, project, parentId)
		if !sameBranch {
			index++
		}
		return index, err
	}

	index = targetTask.Index
	down := sameBranch && targetTask.Index > task.Index

	if !reverse && !down {
		index++
	}

	if reverse && down {
		index--
	}

	return index, nil
}

func (s *tasks) createTempTask(dbCtx *data.DBContext, temp TempTask, parent, index int) (id int, err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	t := temp.Task
	t.ID = 0
	t.ParentID = common.TID(parent)
	t.Index = index

	id, err = s.store.Tasks.Add(dbCtx, &t)
	return id, err
}
